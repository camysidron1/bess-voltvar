package api

import "gopkg.in/yaml.v3"

func yamlUnmarshalImpl(b []byte, out any) error {
	return yaml.Unmarshal(b, out)
}
