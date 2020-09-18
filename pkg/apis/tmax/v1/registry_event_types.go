package v1

type RegistryEvents struct {
	Events []RegistryEvent `json:"events"`
}

type RegistryEvent struct {
	Id             string             `json:"id"`
	Timestamp      string             `json:"timestamp"`
	Action         string             `json:"action"`
	Target         RegistryDescriptor `json:"target"`
	Length         int                `json:"length"`
	Repository     string             `json:"repository"`
	FromRepository string             `json:"fromRepository"`
	Url            string             `json:"url"`
	Tag            string             `json:"tag"`
	Request        RequestRecord      `json:"request"`
	Actor          ActorRecord        `json:"actor"`
	Source         SourceRecord       `json:"source"`
}

type RegistryDescriptor struct {
	MediaType  string `json:"mediaType"`
	Size       int    `json:"size"`
	Digest     string `json:"digest"`
	Length     int    `json:"length"`
	Repository string `json:"repository"`
	Url        string `json:"url"`
	Tag        string `json:"tag"`
}

type RequestRecord struct {
	Id        string `json:"id"`
	Addr      string `json:"addr"`
	Host      string `json:"host"`
	Method    string `json:"method"`
	Useragent string `json:"useragent"`
}

type ActorRecord struct {
	Name string `json:"name"`
}

type SourceRecord struct {
	Addr       string `json:"addr"`
	InstanceID string `json:"instanceID"`
}
