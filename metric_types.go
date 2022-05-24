package main

// Used for choose received event type
type mappingStruct struct {
	HttpMethod string `json:"httpMethod,omitempty"`
	Jvm        string `json:"jwm,omitempty"`
}

type HttpEvent struct {
	HttpMethod            string `json:"httpMethod,omitempty"`
	Status                int    `json:"status,omitempty"`
	Host                  string `json:"host,omitempty"`
	URI                   string `json:"uri,omitempty"`
	Path                  string `json:"path,omitempty"`
	Api                   string `json:"api,omitempty"`
	Application           string `json:"application,omitempty"`
	ApiResponseTimeMs     int    `json:"apiResponseTimeMs,omitempty"`
	RequestContentLength  int    `json:"requestContentLength,omitempty"`
	ResponseContentLength int    `json:"responseContentLength,omitempty"`
	ProxyLatencyMs        int    `json:"proxyLatencyMs,omitempty"`
}
