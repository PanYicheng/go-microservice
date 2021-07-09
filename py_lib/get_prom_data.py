from prometheus_api_client import PrometheusConnect
import datetime
import numpy as np

def get_query1(service):
    return "rate({service}{}) / rate({service}{})".format(
                  '_Index_histogram_sum{service_histogram="duration"}[1m]',
                  '_Index_histogram_count{service_histogram="duration"}[1m]', service=service)

def get_query2(service):
    return service+"_Index_summary{service_summary=\"duration\"}"

def query_prom_data_range(svc_names, query_fn, start_time, end_time, sampling_rate=1, is_summary=False, url="http://vmhost1.local:9090"):
    """Query Prometheus metric data for customized services during customized time range.
    
    Params:
        svc_names: service metric names
        query_fn: function to construct the Prometheus query string from the service name.
        start_time: start time. A datetime.datetime object.
        end_time: same as start. A datetime.datetime object.
        sampling_rate: float, in seconds.
        is_summary: Boolean to represent whether the query is a summary with quantiles.
    
    Returns:
        all_metric_data: A dict of all metric data. Keys are service names. 
            Values are dict containing timestamps and values (If is_summary is True, there are multiple timestamp and value items).
    """
    
    def append_data(d, key, l):
        if key in d:
            d[key].append(l)
        else:
            d[key] = [l]
            
    prom = PrometheusConnect(url = url, disable_ssl=True)
    all_metric_data = {}
    for n in svc_names:
        query = query_fn(n)

        # Split into 3-hour batch and get one batch at a time.
        batch_len = datetime.timedelta(hours=3)
        batch_start = start_time
        batch_end = start_time + batch_len
        timestamps_dict = {}
        values_dict = {}
        metric_info = None
        while batch_start < end_time:
            if batch_end >= end_time:
                batch_end = end_time
            metric_data = prom.custom_query_range(
                        query=query,
                        start_time=batch_start,
                        end_time=batch_end,
                        step=sampling_rate)
            # Sometimes there are no metric data within the range. Skip processing.
            if len(metric_data) > 0:
                if metric_info is None:
                    metric_info = {}
                    metric_info['metric'] = metric_data[0]['metric'].copy()

                for one_data in metric_data:
                    raw_values = np.array(one_data['values'], dtype=np.float64)
                    # Retrive multiple time series data for different quantiles.
                    if is_summary is True:
                        # Remove quantile from metric info.
                        metric_info['metric'].pop('quantile', None)
                        key = 'q'+one_data['metric']['quantile']
                    else:
                        # Only one time series
                        key = 'data'
                    append_data(timestamps_dict, key, raw_values[:, 0])                 
                    append_data(values_dict, key, raw_values[:, 1])

            # Because the previous range [batch_start, batch_end] is inclusive at both ends.
            # We move to the next timestamp here.
            batch_start = batch_end + datetime.timedelta(seconds=sampling_rate)
            batch_end = batch_start + batch_len

        def concat(d, name, conv_type=np.float64):
            for k, v in d.items():
                merged_v = np.concatenate(v).astype(conv_type)
                metric_info[f'{name}_{k}'] = merged_v
        concat(timestamps_dict, 'timestamps', conv_type=np.int64)
        concat(values_dict, 'values')
        all_metric_data[n] = metric_info
    return all_metric_data