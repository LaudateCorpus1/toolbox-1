package toolbox_test

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"
)

type ReverseService struct{}

func (this ReverseService) Reverse(values []int) []int {
	var result = make([]int, 0)
	for i := len(values) - 1; i >= 0; i-- {
		result = append(result, values[i])
	}

	return result
}

func (this ReverseService) Reverse2(values []int) []int {
	var result = make([]int, 0)
	for i := len(values) - 1; i >= 0; i-- {
		result = append(result, values[i])
	}
	return result
}

var ReverseInvoker = func(serviceRouting *toolbox.ServiceRouting, request *http.Request, response http.ResponseWriter, uriParameters map[string]interface{}) error {
	var function = serviceRouting.Handler.(func(values []int) []int)
	idsParam := uriParameters["ids"]
	ids := idsParam.([]string)
	values := make([]int, 0)
	for _, item := range ids {
		values = append(values, toolbox.AsInt(item))
	}
	var result = function(values)
	err := toolbox.WriteServiceRoutingResponse(response, request, serviceRouting, result)
	if err != nil {
		return err
	}
	return nil
}

func StartServer(port string, t *testing.T) {
	service := ReverseService{}
	router := toolbox.NewServiceRouter(
		toolbox.ServiceRouting{
			HTTPMethod: "GET",
			URI:        "/v1/reverse/{ids}",
			Handler:    service.Reverse,
			Parameters: []string{"ids"},
		},
		toolbox.ServiceRouting{
			HTTPMethod: "POST",
			URI:        "/v1/reverse/",
			Handler:    service.Reverse,
			Parameters: []string{"ids"},
		},
		toolbox.ServiceRouting{
			HTTPMethod: "DELETE",
			URI:        "/v1/delete/{ids}",
			Handler:    service.Reverse,
			Parameters: []string{"ids"},
		},
		toolbox.ServiceRouting{
			HTTPMethod:     "GET",
			URI:            "/v1/reverse2/{ids}",
			Handler:        service.Reverse,
			Parameters:     []string{"ids"},
			HandlerInvoker: ReverseInvoker,
		},
		toolbox.ServiceRouting{
			HTTPMethod: "GET",
			URI:        "/v1/tasks",
			Parameters: []string{"status"},
			Handler: func(status string) map[string]interface{} {
				var result = map[string]interface{}{
					"STATUS": status,
					"ABc":    101,
				}
				return result
			},
		},
		toolbox.ServiceRouting{
			HTTPMethod: "GET",
			URI:        "/v1/tasks/{ids}",
			Parameters: []string{"ids"},
			Handler: func(ids ...string) map[string]interface{} {
				var result = map[string]interface{}{
					"STATUS": ids,
					"ABc":    102,
				}
				return result
			},
		},
		toolbox.ServiceRouting{
			HTTPMethod: "GET",
			URI:        "/v1/explain/{cluster}/{yyyy}/{mm}/{dd}/{hh}/{name}/{sid}/{aid}",
			Parameters: []string{"cluster", "sid"},
			Handler: func(amap map[string]interface{}) map[string]interface{} {
				return amap
			},
			HandlerInvoker: func(serviceRouting *toolbox.ServiceRouting, request *http.Request, response http.ResponseWriter, parameters map[string]interface{}) error {
				var result = make(map[string]interface{})
				for k, v := range parameters {
					if strings.HasPrefix(k, "@") {
						continue
					}
					result[k] = v
				}
				data, _ := json.Marshal(result)
				response.Write(data)
				return nil
			},
		},
	)

	http.HandleFunc("/v1/", func(writer http.ResponseWriter, reader *http.Request) {
		err := router.Route(writer, reader)
		assert.Nil(t, err)
	})

	fmt.Printf("Started test server on port %v\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func TestServiceRouter(t *testing.T) {
	go func() {
		StartServer("8082", t)
	}()

	time.Sleep(2 * time.Second)

	{ //Test explain
		var result = map[string]interface{}{}
		err := toolbox.RouteToService("get", "http://127.0.0.1:8082/v1/explain/west/2020/07/26/22/22-02.0.i-08b66c734e4d10a60/59af369e-cf8b-11ea-8858-134d48b29f18/1258652", nil, &result)
		if err != nil {
			t.Errorf("failed to send explain request  %v", err)
		}
		assert.EqualValues(t, map[string]interface{}{
			"cluster": "west",
			"sid": "59af369e-cf8b-11ea-8858-134d48b29f18",
		}, result)
	}

	{

		var result = map[string]interface{}{}
		err := toolbox.RouteToService("get", "http://127.0.0.1:8082/v1/tasks/1,2,3", nil, &result)
		if err != nil {
			t.Errorf("failed to send get request  %v", err)
		}

		assert.EqualValues(t, []interface{}{"1", "2", "3"}, result["STATUS"])

	}

	//
	{

		var result = map[string]interface{}{}
		err := toolbox.RouteToService("get", "http://127.0.0.1:8082/v1/tasks?status=OK", nil, &result)
		if err != nil {
			t.Errorf("failed to send get request  %v", err)
		}

		assert.EqualValues(t, "OK", result["STATUS"])

	}

	var result = make([]int, 0)

	{

		err := toolbox.RouteToService("get", "http://127.0.0.1:8082/v1/reverse/1,7,3", nil, &result)
		if err != nil {
			t.Errorf("failed to send get request  %v", err)
		}
		assert.EqualValues(t, []int{3, 7, 1}, result)

	}

	{

		err := toolbox.RouteToService("post", "http://127.0.0.1:8082/v1/reverse/", []int{1, 7, 3}, &result)
		if err != nil {
			t.Errorf("failed to send get request  %v", err)
		}
		assert.EqualValues(t, []int{3, 7, 1}, result)
	}
	{

		err := toolbox.RouteToService("delete", "http://127.0.0.1:8082/v1/delete/", []int{1, 7, 3}, &result)
		if err != nil {
			t.Errorf("failed to send delete request  %v", err)
		}
		assert.EqualValues(t, []int{3, 7, 1}, result)
	}
	{

		err := toolbox.RouteToService("delete", "http://127.0.0.1:8082/v1/delete/1,7,3", nil, &result)
		if err != nil {
			t.Errorf("failed to send delete request  %v", err)
		}
		assert.EqualValues(t, []int{3, 7, 1}, result)
	}

	{ //Test custom handler invocation without reflection

		err := toolbox.RouteToService("get", "http://127.0.0.1:8082/v1/reverse2/1,7,3", nil, &result)
		if err != nil {
			t.Errorf("failed to send delete request  %v", err)
		}
		assert.EqualValues(t, []int{3, 7, 1}, result)
	}

	{ //Test custom handler invocation without reflection

		err := toolbox.RouteToService("get", "http://127.0.0.1:8082/v1/reverse2/1,7,3", nil, &result)
		if err != nil {
			t.Errorf("failed to send delete request  %v", err)
		}
		assert.EqualValues(t, []int{3, 7, 1}, result)
	}

}
