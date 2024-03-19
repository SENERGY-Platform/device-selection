package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/device-selection/pkg"
	"github.com/SENERGY-Platform/device-selection/pkg/configuration"
	"github.com/SENERGY-Platform/device-selection/pkg/model"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestSNRGY3207(t *testing.T) {
	t.Skip("kubectl port-forward -n permissions deployments/query 8081:8080")
	t.Skip("kubectl port-forward -n device-repository deployments/api 8082:8080")
	config, err := configuration.Load("../../config.json")
	if err != nil {
		t.Error(err)
		return
	}
	config.PermSearchUrl = "http://localhost:8081"
	config.DeviceRepoUrl = "http://localhost:8082"

	ctx, cancel := context.WithCancel(context.Background())
	wg, err := pkg.Start(ctx, config)
	if err != nil {
		t.Error(err)
		return
	}
	defer wg.Wait()
	defer cancel()

	time.Sleep(time.Second)

	endpoint := `/v2/query/selectables?include_groups=false&include_imports=false&include_devices=true&include_id_modified=true&local_devices=A01012124000921:3,A01012124000921:4`
	body := `[{"interaction":"","function_id":"urn:infai:ses:measuring-function:57dfd369-92db-462c-aca4-a767b52c972e","aspect_id":"urn:infai:ses:aspect:74a7b913-73ac-42b7-9b35-573f2c1e97cf","device_class_id":""}]`
	//user ca4d1149-e3ed-4e0b-9e49-3bda908de436; test token will be invalid
	token := `Bearer eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICIzaUtabW9aUHpsMmRtQnBJdS1vSkY4ZVVUZHh4OUFIckVOcG5CcHM5SjYwIn0.eyJleHAiOjE3MTA4NDU3NzQsImlhdCI6MTcxMDg0MjE3NCwianRpIjoiNTRlNGI0NmItN2E5ZS00YTJmLWE4ZmItNmRiMjlkZWMyZGFlIiwiaXNzIjoiaHR0cHM6Ly9hdXRoLnNlbmVyZ3kuaW5mYWkub3JnL2F1dGgvcmVhbG1zL21hc3RlciIsImF1ZCI6ImFjY291bnQiLCJzdWIiOiJjYTRkMTE0OS1lM2VkLTRlMGItOWU0OS0zYmRhOTA4ZGU0MzYiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJmcm9udGVuZCIsIm5vbmNlIjoiNTRkNjYxZTItYjhhOC00MTIwLTg5ODYtN2YxNzkwNThlZGRkIiwic2Vzc2lvbl9zdGF0ZSI6IjhiNGIyODc4LTVhZjEtNDE5MC04MDA4LTU1NWE1YzQxNzIwZCIsImFsbG93ZWQtb3JpZ2lucyI6WyIqIl0sInJlYWxtX2FjY2VzcyI6eyJyb2xlcyI6WyJmdWxsX3VzZXIiLCJvZmZsaW5lX2FjY2VzcyIsImRldmVsb3BlciIsInVtYV9hdXRob3JpemF0aW9uIiwidXNlciJdfSwicmVzb3VyY2VfYWNjZXNzIjp7ImFjY291bnQiOnsicm9sZXMiOlsibWFuYWdlLWFjY291bnQiLCJtYW5hZ2UtYWNjb3VudC1saW5rcyIsInZpZXctcHJvZmlsZSJdfX0sInNjb3BlIjoib3BlbmlkIHByb2ZpbGUgZW1haWwiLCJzaWQiOiI4YjRiMjg3OC01YWYxLTQxOTAtODAwOC01NTVhNWM0MTcyMGQiLCJlbWFpbF92ZXJpZmllZCI6ZmFsc2UsInJvbGVzIjpbImZ1bGxfdXNlciIsIm9mZmxpbmVfYWNjZXNzIiwiZGV2ZWxvcGVyIiwidW1hX2F1dGhvcml6YXRpb24iLCJ1c2VyIl0sInByZWZlcnJlZF91c2VybmFtZSI6ImpvbmFoIiwibG9jYWxlIjoiZGUifQ.KsO9DwEvRkSp8LQGf6Vwrd1SAQMlfdV2cLVezGmOnMUlQnPfhOiUE7QczbWcEeYcRAfk4UPNBUHtbATFv9AJ7G891XGzL8mVrOMHVn2AFL_3vAbZn6J1gWmi4HxTxOCHiWuZoDY_BroV0bstsbVOOXt_7OMwv55DaKjTb03Esgp1oruuB8bnBgsgFnRkeWkWyoYzMo_p83CcXEB8jmXoUM1uGArf6RAX2U4PQDMUGGyCF1QJ7pZWnjmQZq2iUSXvW6Wnh8Y4scJ2l7w8GbeSUCceKdxmPdES10qYlD-VPfGwuxHYZzx7_gOU8loXs5lo1G9AwlCuN1pZmap4OTS-rA`

	expectedResponseJson := `[{"device":{"id":"urn:infai:ses:device:f5543003-7811-44ab-8d13-bd592958ffa4","local_id":"A01012124000921:3","name":"Plug Waschmaschine","attributes":[{"key":"optimise-mobile-app/favorite","value":"true","origin":"optimise-mobile-app"}],"device_type_id":"urn:infai:ses:device-type:4450cb9c-a6f7-4962-9d1d-70cc2995f36c","display_name":"Plug Waschmaschine","permissions":{"r":true,"w":true,"x":true,"a":true},"shared":false,"creator":"ca4d1149-e3ed-4e0b-9e49-3bda908de436"},"services":[{"id":"urn:infai:ses:service:f171a0b5-0bb1-438f-8ea8-720b156b13a1","local_id":"50-0-value-65537:get","name":"Get Electricity Consumption","description":"","interaction":"event+request","protocol_id":"urn:infai:ses:protocol:f3a63aeb-187e-4dd9-9ef5-d97a6eb6292b","inputs":[],"outputs":[{"id":"urn:infai:ses:content:2be10729-6775-4539-856e-2243249c9007","content_variable":{"id":"urn:infai:ses:content-variable:e0a722df-9fa8-415f-8649-c5fa4eadf8e1","name":"root","is_void":false,"type":"https://schema.org/StructuredValue","sub_content_variables":[{"id":"urn:infai:ses:content-variable:4ec693e6-cd16-4ded-977d-0d90d255e53c","name":"value","is_void":false,"type":"https://schema.org/Float","sub_content_variables":null,"characteristic_id":"urn:infai:ses:characteristic:3febed55-ba9b-43dc-8709-9c73bae3716e","value":null,"serialization_options":null,"function_id":"urn:infai:ses:measuring-function:57dfd369-92db-462c-aca4-a767b52c972e","aspect_id":"urn:infai:ses:aspect:74a7b913-73ac-42b7-9b35-573f2c1e97cf"},{"id":"urn:infai:ses:content-variable:74271ce2-eb3b-47cf-8599-0a4a3fdd6d40","name":"lastUpdate","is_void":false,"type":"https://schema.org/Integer","sub_content_variables":null,"characteristic_id":"urn:infai:ses:characteristic:64691f8d-4909-470f-a1fa-e977ebe28684","value":null,"serialization_options":null,"function_id":"urn:infai:ses:measuring-function:3b4e0766-0d67-4658-b249-295902cd3290","aspect_id":"urn:infai:ses:aspect:74a7b913-73ac-42b7-9b35-573f2c1e97cf"},{"id":"urn:infai:ses:content-variable:c63570ca-4e06-494b-8b97-8ae3e6ed863b","name":"value_unit","is_void":false,"type":"https://schema.org/Text","sub_content_variables":null,"characteristic_id":"","value":null,"serialization_options":null,"unit_reference":"value"},{"id":"urn:infai:ses:content-variable:7e2b776c-2eac-4a4e-8e4b-471434485db4","name":"lastUpdate_unit","is_void":false,"type":"https://schema.org/Text","sub_content_variables":null,"characteristic_id":"","value":null,"serialization_options":null,"unit_reference":"lastUpdate"}],"characteristic_id":"","value":null,"serialization_options":null},"serialization":"json","protocol_segment_id":"urn:infai:ses:protocol-segment:0d211842-cef8-41ec-ab6b-9dbc31bc3a65"}],"attributes":[],"service_group_key":""}],"servicePathOptions":{"urn:infai:ses:service:f171a0b5-0bb1-438f-8ea8-720b156b13a1":[{"path":"root.value","characteristicId":"urn:infai:ses:characteristic:3febed55-ba9b-43dc-8709-9c73bae3716e","aspectNode":{"id":"urn:infai:ses:aspect:74a7b913-73ac-42b7-9b35-573f2c1e97cf","name":"Consumption","root_id":"urn:infai:ses:aspect:412a48ad-3a80-46f7-8b99-408c4b9c3528","parent_id":"urn:infai:ses:aspect:412a48ad-3a80-46f7-8b99-408c4b9c3528","child_ids":["urn:infai:ses:aspect:34fdb3f7-5366-4746-82a1-89b0ca90263f","urn:infai:ses:aspect:4e110c6a-6cdd-4a25-9eb9-dbeec5e3c22c","urn:infai:ses:aspect:82b9e844-1fa6-4fb5-8213-ec7333a9a8ba","urn:infai:ses:aspect:b58c91a8-0db3-44df-a66e-94d438dea97e","urn:infai:ses:aspect:c407cb46-01a3-49b5-997d-25aa343fff6f","urn:infai:ses:aspect:c9ea9ed3-14e2-4a0d-aa41-668be5b1fd50","urn:infai:ses:aspect:d0385a96-d9e7-45e0-8c0d-df6e97a8da05","urn:infai:ses:aspect:e878a59a-248f-440e-b631-ccf1af8210f2","urn:infai:ses:aspect:fdc999eb-d366-44e8-9d24-bfd48d5fece1"],"ancestor_ids":["urn:infai:ses:aspect:412a48ad-3a80-46f7-8b99-408c4b9c3528"],"descendent_ids":["urn:infai:ses:aspect:34fdb3f7-5366-4746-82a1-89b0ca90263f","urn:infai:ses:aspect:4e110c6a-6cdd-4a25-9eb9-dbeec5e3c22c","urn:infai:ses:aspect:82b9e844-1fa6-4fb5-8213-ec7333a9a8ba","urn:infai:ses:aspect:b58c91a8-0db3-44df-a66e-94d438dea97e","urn:infai:ses:aspect:c407cb46-01a3-49b5-997d-25aa343fff6f","urn:infai:ses:aspect:c9ea9ed3-14e2-4a0d-aa41-668be5b1fd50","urn:infai:ses:aspect:d0385a96-d9e7-45e0-8c0d-df6e97a8da05","urn:infai:ses:aspect:e878a59a-248f-440e-b631-ccf1af8210f2","urn:infai:ses:aspect:fdc999eb-d366-44e8-9d24-bfd48d5fece1"]},"functionId":"urn:infai:ses:measuring-function:57dfd369-92db-462c-aca4-a767b52c972e","isVoid":false,"type":"https://schema.org/Float","interaction":"event+request"}]}},{"device":{"id":"urn:infai:ses:device:2ac5436e-5538-4eb3-a448-2d77de68e915","local_id":"A01012124000921:4","name":"Qubino","attributes":[{"key":"optimise-mobile-app/favorite","value":"true","origin":"optimise-mobile-app"}],"device_type_id":"urn:infai:ses:device-type:877ede3a-26b7-406d-93a9-c40c0dbfed75","display_name":"Qubino","permissions":{"r":true,"w":true,"x":true,"a":true},"shared":false,"creator":"ca4d1149-e3ed-4e0b-9e49-3bda908de436"},"services":[{"id":"urn:infai:ses:service:cfc6a85f-01ab-4017-9d57-92fb050bc0fa","local_id":"50-1-value-65537:get","name":"Get Electricity Consumption - Total","description":"Electric Meter kWh","interaction":"event+request","protocol_id":"urn:infai:ses:protocol:f3a63aeb-187e-4dd9-9ef5-d97a6eb6292b","inputs":[],"outputs":[{"id":"urn:infai:ses:content:237e927b-7d17-49d0-a634-d1f1e1336186","content_variable":{"id":"urn:infai:ses:content-variable:f840924a-f5e9-4d4c-af62-765e69d911ee","name":"energy","is_void":false,"type":"https://schema.org/StructuredValue","sub_content_variables":[{"id":"urn:infai:ses:content-variable:ea5fa8a8-e85a-4686-aa3d-572bf6feb94a","name":"value","is_void":false,"type":"https://schema.org/Float","sub_content_variables":null,"characteristic_id":"urn:infai:ses:characteristic:3febed55-ba9b-43dc-8709-9c73bae3716e","value":null,"serialization_options":null,"function_id":"urn:infai:ses:measuring-function:57dfd369-92db-462c-aca4-a767b52c972e","aspect_id":"urn:infai:ses:aspect:fdc999eb-d366-44e8-9d24-bfd48d5fece1"},{"id":"urn:infai:ses:content-variable:c01c340d-bf62-4d14-a15b-f915601a61c2","name":"lastUpdate","is_void":false,"type":"https://schema.org/Integer","sub_content_variables":null,"characteristic_id":"urn:infai:ses:characteristic:64691f8d-4909-470f-a1fa-e977ebe28684","value":null,"serialization_options":null,"function_id":"urn:infai:ses:measuring-function:3b4e0766-0d67-4658-b249-295902cd3290","aspect_id":"urn:infai:ses:aspect:74a7b913-73ac-42b7-9b35-573f2c1e97cf"},{"id":"urn:infai:ses:content-variable:627aa390-3a24-43d2-a949-11ebb1bcbb1d","name":"value_unit","is_void":false,"type":"https://schema.org/Text","sub_content_variables":null,"characteristic_id":"","value":null,"serialization_options":null,"unit_reference":"value"},{"id":"urn:infai:ses:content-variable:62b3e39e-bdfc-49fc-b528-a67412f274cd","name":"lastUpdate_unit","is_void":false,"type":"https://schema.org/Text","sub_content_variables":null,"characteristic_id":"","value":null,"serialization_options":null,"unit_reference":"lastUpdate"}],"characteristic_id":"","value":null,"serialization_options":null},"serialization":"json","protocol_segment_id":"urn:infai:ses:protocol-segment:0d211842-cef8-41ec-ab6b-9dbc31bc3a65"}],"attributes":[{"key":"senergy/time_path","value":"","origin":"web-ui"}],"service_group_key":"c3b8fc33-1899-4595-a6fc-eb85ff24a0de"}],"servicePathOptions":{"urn:infai:ses:service:cfc6a85f-01ab-4017-9d57-92fb050bc0fa":[{"path":"energy.value","characteristicId":"urn:infai:ses:characteristic:3febed55-ba9b-43dc-8709-9c73bae3716e","aspectNode":{"id":"urn:infai:ses:aspect:fdc999eb-d366-44e8-9d24-bfd48d5fece1","name":"Total","root_id":"urn:infai:ses:aspect:412a48ad-3a80-46f7-8b99-408c4b9c3528","parent_id":"urn:infai:ses:aspect:74a7b913-73ac-42b7-9b35-573f2c1e97cf","child_ids":[],"ancestor_ids":["urn:infai:ses:aspect:412a48ad-3a80-46f7-8b99-408c4b9c3528","urn:infai:ses:aspect:74a7b913-73ac-42b7-9b35-573f2c1e97cf"],"descendent_ids":[]},"functionId":"urn:infai:ses:measuring-function:57dfd369-92db-462c-aca4-a767b52c972e","isVoid":false,"type":"https://schema.org/Float","interaction":"event+request"}]}},{"device":{"id":"urn:infai:ses:device:2ac5436e-5538-4eb3-a448-2d77de68e915$service_group_selection=c3b8fc33-1899-4595-a6fc-eb85ff24a0de","local_id":"A01012124000921:4","name":"Qubino Total","attributes":[{"key":"optimise-mobile-app/favorite","value":"true","origin":"optimise-mobile-app"}],"device_type_id":"urn:infai:ses:device-type:877ede3a-26b7-406d-93a9-c40c0dbfed75$service_group_selection=c3b8fc33-1899-4595-a6fc-eb85ff24a0de","display_name":"Qubino Total","permissions":{"r":true,"w":true,"x":true,"a":true},"shared":false,"creator":"ca4d1149-e3ed-4e0b-9e49-3bda908de436"},"services":[{"id":"urn:infai:ses:service:cfc6a85f-01ab-4017-9d57-92fb050bc0fa","local_id":"50-1-value-65537:get","name":"Get Electricity Consumption - Total","description":"Electric Meter kWh","interaction":"event+request","protocol_id":"urn:infai:ses:protocol:f3a63aeb-187e-4dd9-9ef5-d97a6eb6292b","inputs":[],"outputs":[{"id":"urn:infai:ses:content:237e927b-7d17-49d0-a634-d1f1e1336186","content_variable":{"id":"urn:infai:ses:content-variable:f840924a-f5e9-4d4c-af62-765e69d911ee","name":"energy","is_void":false,"type":"https://schema.org/StructuredValue","sub_content_variables":[{"id":"urn:infai:ses:content-variable:ea5fa8a8-e85a-4686-aa3d-572bf6feb94a","name":"value","is_void":false,"type":"https://schema.org/Float","sub_content_variables":null,"characteristic_id":"urn:infai:ses:characteristic:3febed55-ba9b-43dc-8709-9c73bae3716e","value":null,"serialization_options":null,"function_id":"urn:infai:ses:measuring-function:57dfd369-92db-462c-aca4-a767b52c972e","aspect_id":"urn:infai:ses:aspect:fdc999eb-d366-44e8-9d24-bfd48d5fece1"},{"id":"urn:infai:ses:content-variable:c01c340d-bf62-4d14-a15b-f915601a61c2","name":"lastUpdate","is_void":false,"type":"https://schema.org/Integer","sub_content_variables":null,"characteristic_id":"urn:infai:ses:characteristic:64691f8d-4909-470f-a1fa-e977ebe28684","value":null,"serialization_options":null,"function_id":"urn:infai:ses:measuring-function:3b4e0766-0d67-4658-b249-295902cd3290","aspect_id":"urn:infai:ses:aspect:74a7b913-73ac-42b7-9b35-573f2c1e97cf"},{"id":"urn:infai:ses:content-variable:627aa390-3a24-43d2-a949-11ebb1bcbb1d","name":"value_unit","is_void":false,"type":"https://schema.org/Text","sub_content_variables":null,"characteristic_id":"","value":null,"serialization_options":null,"unit_reference":"value"},{"id":"urn:infai:ses:content-variable:62b3e39e-bdfc-49fc-b528-a67412f274cd","name":"lastUpdate_unit","is_void":false,"type":"https://schema.org/Text","sub_content_variables":null,"characteristic_id":"","value":null,"serialization_options":null,"unit_reference":"lastUpdate"}],"characteristic_id":"","value":null,"serialization_options":null},"serialization":"json","protocol_segment_id":"urn:infai:ses:protocol-segment:0d211842-cef8-41ec-ab6b-9dbc31bc3a65"}],"attributes":[{"key":"senergy/time_path","value":"","origin":"web-ui"}],"service_group_key":"c3b8fc33-1899-4595-a6fc-eb85ff24a0de"}],"servicePathOptions":{"urn:infai:ses:service:cfc6a85f-01ab-4017-9d57-92fb050bc0fa":[{"path":"energy.value","characteristicId":"urn:infai:ses:characteristic:3febed55-ba9b-43dc-8709-9c73bae3716e","aspectNode":{"id":"urn:infai:ses:aspect:fdc999eb-d366-44e8-9d24-bfd48d5fece1","name":"Total","root_id":"urn:infai:ses:aspect:412a48ad-3a80-46f7-8b99-408c4b9c3528","parent_id":"urn:infai:ses:aspect:74a7b913-73ac-42b7-9b35-573f2c1e97cf","child_ids":[],"ancestor_ids":["urn:infai:ses:aspect:412a48ad-3a80-46f7-8b99-408c4b9c3528","urn:infai:ses:aspect:74a7b913-73ac-42b7-9b35-573f2c1e97cf"],"descendent_ids":[]},"functionId":"urn:infai:ses:measuring-function:57dfd369-92db-462c-aca4-a767b52c972e","isVoid":false,"type":"https://schema.org/Float","interaction":"event+request"}]}}]`

	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080"+endpoint, bytes.NewBuffer([]byte(body)))
	if err != nil {
		t.Error(err)
		return
	}
	req.Header.Set("Authorization", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != 200 {
		t.Error(resp.StatusCode)
		return
	}

	expectedSelectables := []model.Selectable{}
	err = json.Unmarshal([]byte(expectedResponseJson), &expectedSelectables)
	if err != nil {
		t.Error(err)
		return
	}
	expectedSelectables = normalize(expectedSelectables)

	actualSelectables := []model.Selectable{}
	err = json.NewDecoder(resp.Body).Decode(&actualSelectables)
	if err != nil {
		t.Error(err)
		return
	}
	actualSelectables = normalize(actualSelectables)

	if !reflect.DeepEqual(actualSelectables, expectedSelectables) {
		t.Errorf("\n%#v\n%#v\n", actualSelectables, expectedSelectables)
	}
}

func normalize(selectables []model.Selectable) []model.Selectable {
	slices.SortFunc(selectables, func(a, b model.Selectable) int {
		return strings.Compare(a.Device.Id, b.Device.Id)
	})
	result := []model.Selectable{}
	for _, val := range selectables {
		for key, options := range val.ServicePathOptions {
			slices.SortFunc(options, func(a, b model.PathOption) int {
				return strings.Compare(a.Path, b.Path)
			})
			val.ServicePathOptions[key] = options
		}
		result = append(result, val)
	}
	return result
}
