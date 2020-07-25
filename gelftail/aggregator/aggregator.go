package aggregator

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

var client = &http.Client{}
var logglyBaseUrl = "https://logs-01.loggly.com/inputs/%s/tag/http/"
var url string

func Start(bulkQueue chan []byte, authToken string) {
	url = fmt.Sprintf(logglyBaseUrl, authToken) // Assemble the final loggly bulk upload URL using the authToken
	buf := new(bytes.Buffer)
	for {
		msg := <-bulkQueue // Blocks here until a message arrives on the channel.
		buf.Write(msg)
		buf.WriteString("\n") // Loggly needs newline to separate log statements properly.

		size := buf.Len()
		if size > 1024 { // If buffer has more than 1024 bytes of data...
			sendBulk(*buf) // Upload!
			buf.Reset()
		}
	}
}

func sendBulk(buffer bytes.Buffer) {
	req, err := http.NewRequest("POST", url,
		bytes.NewReader(buffer.Bytes()))

	if err != nil {
		logrus.Errorln("Error creating bulk upload HTTP request: " + err.Error())
		return
	}
	resp, err := client.Do(req)

	if err != nil || resp.StatusCode != 200 {
		logrus.Errorln("Error sending bulk: " + err.Error())
		return
	}
	logrus.Debugf("Successfully sent batch of %v bytes to Loggly\n", buffer.Len())
}
