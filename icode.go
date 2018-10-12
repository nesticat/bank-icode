/*
 * Copyright 2018 It-chain
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// should not modify this file
// this file will be used as just test

package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"encoding/json"

	"github.com/it-chain/sdk"
	"github.com/it-chain/sdk/logger"
	"github.com/it-chain/sdk/pb"
	"github.com/jessevdk/go-flags"
)

var opts struct {
	Port int `short:"p" long:"port" description:"set port"`
}

func main() {

	logger.EnableFileLogger(true, "./icode.log")
	parser := flags.NewParser(&opts, flags.Default)

	_, err := parser.Parse()

	if err != nil {
		logger.Error(nil, "fail parse args: "+err.Error())
		os.Exit(1)
	}

	fmt.Println("port : " + strconv.Itoa(opts.Port))

	exHandler := &HandlerExample{}
	ibox := sdk.NewIBox(opts.Port)
	ibox.SetHandler(exHandler)
	err = ibox.On(30)

	if err != nil {
		panic(err.Error())
	}
}

type HandlerExample struct {
}

func (*HandlerExample) Name() string {
	return "sample"
}

func (*HandlerExample) Versions() []string {
	vers := make([]string, 0)
	vers = append(vers, "1.0")
	vers = append(vers, "1.2")
	return vers
}

func (*HandlerExample) Handle(request *pb.Request, cell *sdk.Cell) *pb.Response {
	switch request.Type {
	case "invoke":
		return handleInvoke(request, cell)
	case "query":
		return handleQuery(request, cell)
	case "test":
		fmt.Println("req : " + request.Uuid)
		if request.Uuid == "0" {
			cell.PutData("test", []byte("0"))
			return responseSuccess(request, []byte(string(0)))
		}
		data, err := cell.GetData("test")
		if err != nil {
			return responseError(request, err)
		}
		if len(data) == 0 {
			err := errors.New("no data err")
			return responseError(request, err)
		}
		strData := string(data)
		intData, err := strconv.Atoi(strData)
		if err != nil {
			return responseError(request, err)
		}
		intData = intData + 1
		changeData := strconv.Itoa(intData)
		err = cell.PutData("test", []byte(changeData))
		if err != nil {
			return responseError(request, err)
		}
		return responseSuccess(request, []byte(changeData))
	default:
		logger.Debug(nil, "unknown request type")
		err := errors.New("unknown request type")
		return responseError(request, err)
	}
}
func handleQuery(request *pb.Request, cell *sdk.Cell) *pb.Response {
	args := request.GetArgs()
	switch request.FunctionName {
	case "accounts":
		it := cell.DBHandler.GetIteratorWithPrefix()
		for it.First(); it.Valid(); it.Next() {
			logger.Error(nil, "query - accounts "+string(it.Key())+"/"+string(it.Value()))
		}

		return responseSuccess(request, nil)
	case "balance":
		b, err := cell.GetData(args[0])
		logger.Error(nil, "query - balance "+args[0]+"/"+string(b))

		if err != nil {
			return responseError(request, err)
		}

		result := make(map[string]string)
		result[args[0]] = string(b)

		d, err := json.Marshal(result)

		if err != nil {
			return responseError(request, err)
		}

		return responseSuccess(request, d)

	default:
		err := errors.New("unknown query method")
		return responseError(request, err)
	}
}
func handleInvoke(request *pb.Request, cell *sdk.Cell) *pb.Response {
	args := request.GetArgs()
	switch request.FunctionName {
	case "mint":
		err := cell.PutData(args[0], []byte(args[1]))
		logger.Error(nil, "invoke - mint "+args[0]+"/"+args[1])
		if err != nil {
			return responseError(request, err)
		}
		return responseSuccess(request, nil)
	case "transfer":
		logger.Error(nil, "invoke - transfer "+args[0]+"/"+args[1]+":"+args[2])
		transBalance, err := strconv.Atoi(args[2])
		if err != nil {
			return responseError(request, err)
		}
		// get balance 1
		data1, err := cell.GetData(args[0])
		if err != nil {
			return responseError(request, err)
		}
		if len(data1) == 0 {
			err := errors.New("no data err")
			return responseError(request, err)
		}
		strData1 := string(data1)
		intData1, err := strconv.Atoi(strData1)
		if err != nil {
			return responseError(request, err)
		}

		// get balance 2
		data2, err := cell.GetData(args[1])
		if err != nil {
			return responseError(request, err)
		}
		if len(data2) == 0 {
			err := errors.New("no data err")
			return responseError(request, err)
		}
		strData2 := string(data2)
		intData2, err := strconv.Atoi(strData2)
		if err != nil {
			return responseError(request, err)
		}

		// put 1,2
		intData1 -= transBalance
		intData2 += transBalance
		changeData1 := strconv.Itoa(intData1)
		err = cell.PutData(args[0], []byte(changeData1))
		if err != nil {
			return responseError(request, err)
		}
		changeData2 := strconv.Itoa(intData2)
		err = cell.PutData(args[1], []byte(changeData2))
		if err != nil {
			return responseError(request, err)
		}
		return responseSuccess(request, nil)
	default:
		err := errors.New("unknown invoke method")
		return responseError(request, err)
	}
}

func responseError(request *pb.Request, err error) *pb.Response {
	return &pb.Response{
		Uuid:  request.Uuid,
		Type:  request.Type,
		Data:  nil,
		Error: err.Error(),
	}
}

func responseSuccess(request *pb.Request, data []byte) *pb.Response {
	return &pb.Response{
		Uuid:  request.Uuid,
		Type:  request.Type,
		Data:  data,
		Error: "",
	}
}
