/*
 *  Copyright (C) 2016 VSCT
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 */
package internal

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"time"
)

type HapInstallation struct {
	Path string
}

type Config struct {
	LookupdAddresses []string
	ProducerAddr     string
	ProducerRestAddr string
	ClusterId        string
	Port             int32
	HapHome          string
	Id               string
	Hap              map[string]HapInstallation
}

func (config Config) IsMaster(vip string) (bool, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s:%d/id", vip, config.Port))
	if err != nil {
		log.WithField("vip", vip).WithField("port", config.Port).WithError(err).Error("can't request in http the vip")
		return false, err
	} else {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		isMaster := string(body) == config.Id
		if log.GetLevel() == log.DebugLevel {
			log.WithField("id", string(body)).WithField("isMaster", isMaster).Debug("check if this instance is master")
		}
		return string(body) == config.Id, err
	}
}

func DefaultConfig() *Config {
	return &Config{
		Port:      5000,
		HapHome:   "/HOME/hapadm",
		ClusterId: "default-name",
	}
}

type Header struct {
	CorrelationId string `json:"correlationId"`
	Application   string `json:"application"`
	Platform      string `json:"platform"`
	Timestamp     int64  `json:"timestamp"`
	Source        string `json:"source"`
}

type Conf struct {
	Haproxy []byte `json:"haproxy"`
	Syslog  []byte `json:"syslog"`
	Bind    string `json:"bind"`
	Version string `json:"haproxyVersion,omitempty"`
}

// main type for messages
type EventMessage struct {
	Header Header `json:"header"`
}

// type of messages with Conf type additionally
type EventMessageWithConf struct {
	EventMessage
	Conf Conf `json:"conf,omitempty"`
}

// clone an EventMessage with he header, a new source and a new timestamp and the conf
func (eventMessage EventMessageWithConf) CloneWithConf(source string) EventMessageWithConf {
	newMessage := EventMessageWithConf{
		Conf: eventMessage.Conf,
	}
	newMessage.Header = eventMessage.Header
	newMessage.Header.Source = source
	newMessage.Header.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
	return newMessage
}

// retrieve Context from an EventMessage
func (eventMessage EventMessage) Context() Context {
	return Context{
		CorrelationId: eventMessage.Header.CorrelationId,
		Timestamp:     eventMessage.Header.Timestamp,
		Application:   eventMessage.Header.Application,
		Platform:      eventMessage.Header.Platform,
	}
}

// clone an EventMessage with just the header, a new source and a new timestamp
func (eventMessage EventMessage) Clone(source string) EventMessage {
	newMessage := EventMessage{
		Header: eventMessage.Header,
	}
	newMessage.Header.Source = source
	newMessage.Header.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
	return newMessage
}

// context for tracing current process and local processing
type Context struct {
	CorrelationId string `json:"correlationId"`
	Timestamp     int64  `json:"timestamp"`
	Application   string `json:"application"`
	Platform      string `json:"platform"`
}

// update the timestamp of the current context
func (ctx Context) UpdateTimestamp() Context {
	ctx.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
	return ctx
}

// log with context headers
func (ctx Context) Fields(fields log.Fields) *log.Entry {
	fields["correlationId"] = ctx.CorrelationId
	fields["timestamp"] = ctx.UpdateTimestamp().Timestamp
	fields["application"] = ctx.Application
	fields["platform"] = ctx.Platform

	return log.WithFields(fields)
}

type ReloadEvent struct {
	Message *EventMessageWithConf
	F       func(data *EventMessageWithConf) error
}

func (re *ReloadEvent) Execute() error {
	return re.F(re.Message)
}

type EventHandler interface {
	HandleMessage(data *EventMessageWithConf) error
}
type HandlerFunc func(data *EventMessageWithConf) error

func (h HandlerFunc) HandleMessage(m *EventMessageWithConf) error {
	return h(m)
}

type SdkLogger struct {
	Logrus *log.Logger
}

func (sdkLogger SdkLogger) Output(calldepth int, s string) error {
	log.WithField("type", "nsq driver").Info(s)
	return nil
}
