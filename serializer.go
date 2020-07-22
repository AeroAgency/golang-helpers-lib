package helpers

import "encoding/json"

type Serializer struct{}

func (s Serializer) ConvertStructToProto(inputStruct interface{}, outputProto interface{}) error {
	data, err := json.Marshal(inputStruct)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &outputProto)
	if err != nil {
		return err
	}
	return nil
}

func (s Serializer) ConvertProtoToStruct(inputProto interface{}, outputStruct interface{}) error {
	data, err := json.Marshal(inputProto)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &outputStruct)
	if err != nil {
		return err
	}
	return nil
}

func (s Serializer) ConvertJsonProtoTo(data string, outputStruct interface{}) error {
	err := json.Unmarshal([]byte(data), &outputStruct)
	if err != nil {
		return err
	}
	return nil
}