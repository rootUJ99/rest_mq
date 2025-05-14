package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	mrand "math/rand/v2"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	utils "github.com/rootuj/async_queue_with_rest/utils"
	"github.com/valkey-io/valkey-go"

)

const MENU_MAP string = "menu_map" 
const QUANTITY_MAP string = "quantity_map"
const PLACE_ORDER_MAP string = "place_order_map"

type ErrorResponse struct {
	Message string `json:"messaage"`
}

type ValkeyModel struct {
	Client valkey.Client
	Ctx context.Context
}

func NewErrorResponse(err error) ErrorResponse {
	if err != nil {
		return ErrorResponse{ Message: fmt.Sprintf("%v", err)}
	}
	return ErrorResponse{ Message: "Oops something went wrong"}
}

type MenuList struct {
	Itemcode  string `json:"item_code"`
	Itemquantity int`json:"item_quantity"`
	Itemvalue string `json:"item_value"`
} 


type MenuBody struct {
	MenuList [] MenuList `json:"menu_list"`
}

type OrderData map[string][]MenuList


type MenuMapOfs struct {
	MenuMapAvailable map[string]int `json:"menu_available"`
	MenuMapUnavailable map[string]bool `json:"menu_unavailable"`
}
func getRandomString() string{
	random_id := ""
	for i := range 7 {
		if i < 3 {
			randN := mrand.N(25) + 65
			r1 := rune(randN)
			random_id = fmt.Sprintf("%s%s", random_id, string(r1))	
			continue	
			
		}
		randN := mrand.N(9)
		random_id = fmt.Sprintf("%s%d", random_id, randN)	
	}
	return random_id 
}

func JSONEncoderWrap(w io.Writer, jsonData any) error {
	err := json.NewEncoder(w).Encode(jsonData)
	if err != nil {
		return fmt.Errorf("Error occured while encoding the data, %w", err)
	}
	return nil
}

func JSONDecodeWrap(r io.Reader, jsonStruct any) error {
	if jsonStruct == nil {
		return fmt.Errorf("please provide a valid struct with json template")
	}
	err := json.NewDecoder(r).Decode(jsonStruct)

	if err != nil {
		return fmt.Errorf("Error occured while decoding the data, %w", err)
	}
	return nil
}

func main(){
	fmt.Println("---- starting the app ----")	
	client, err := valkey.NewClient(valkey.ClientOption{InitAddress: []string{"127.0.0.1:6379"}})
	utils.HandleErrorWithLog(err)
	defer client.Close()
	// ctx := context.Background()

	m := ValkeyModel{Client: client, Ctx: context.Background()}
	fmt.Println("---- Starting the valkey server ----")
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.AllowContentType("application/json"))
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hellow there"))
	})
	router.Post("/place_order", m.handleOrders)
	router.Post("/add_menu", m.addMenu)
	router.Get("/list_ofs", m.listofs)
	router.Post("/make_ofs", m.makeOfs)
	fmt.Println("---- listning on 6969 ---- nice")	
	fmt.Println("---- listning on 6969 ---- nice")	
	http.ListenAndServe(":6969", router)
}


func (v ValkeyModel) handleOrders(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Got here finally")
	orderId := getRandomString()
	var orderDetails MenuBody
	err := JSONDecodeWrap(r.Body, &orderDetails)

	orderData := OrderData{ orderId : make([]MenuList, 0)}
	orderData[orderId] = append(orderData[orderId], orderDetails.MenuList...)

	jsonDdata, err := json.Marshal(orderData)
	utils.HandleErrorWithLog(err)
	fmt.Printf("%s",jsonDdata)
	err = v.Client.Do(v.Ctx, v.Client.B().JsonSet().Key(PLACE_ORDER_MAP).Path("$").Value(string(jsonDdata)).Build()).Error()
	
	ofsJsonData, err := json.Marshal(orderDetails)
	utils.HandleErrorWithLog(err)

	fmt.Println("marshelled data")
	utils.SenderToQueue(ofsJsonData)
	err = JSONEncoderWrap(w, orderData)
	utils.HandleErrorWithLog(err)

}
func (v ValkeyModel) addMenu (w http.ResponseWriter, r *http.Request) {
	fmt.Printf("getting here %s", "got here")
	var menu MenuBody	
	err := JSONDecodeWrap(r.Body, &menu)
	utils.HandleErrorWithLog(err)
	// v.Client.Do(v.Ctx, v.Client.B().Lpush().Key("MENU").Element("abc").Build()).Error()
	fmt.Printf("%v", menu)
	for _, ele := range menu.MenuList {
		key := ele.Itemcode
		value := ele.Itemvalue
		qnt := strconv.Itoa(ele.Itemquantity)
		fmt.Printf("-----> %s %s", key, value)
		err := v.Client.Do(v.Ctx, v.Client.B().Hsetnx().Key(MENU_MAP).Field(key).Value(value).Build()).Error()
		if err != nil {
			utils.HandleErrorWithLog(err)
			break
		}
		err = v.Client.Do(v.Ctx, v.Client.B().Hsetnx().Key(QUANTITY_MAP).Field(key).Value(qnt).Build()).Error()
		if err != nil {
			utils.HandleErrorWithLog(err)
			break
		}
	}
	err = JSONEncoderWrap(w, menu)
	utils.HandleErrorWithLog(err)
}


// func (v ValkeyModel) listmenu(w http.ResponseWriter, r *http.Request) {
// 	menuMap, err := v.Client.Do(v.Ctx, v.Client.B().Hgetall().Key(MENU_MAP).Build()).AsStrMap()
// 	utils.HandleErrorWithLog(err)
// 	utils.HandleErrorWithLog(err)
// 	menubody := MenuMapBody{MenuMap: menuMap}	
// 	err = JSONEncoderWrap(w, menubody)
// 	utils.HandleErrorWithLog(err)
// }

func (v ValkeyModel) listofs(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("this is list ofs %s", "swagg")
	menuMap, err := v.Client.Do(v.Ctx, v.Client.B().Hgetall().Key(MENU_MAP).Build()).AsStrMap()
	utils.HandleErrorWithLog(err)
	quantityMap, err := v.Client.Do(v.Ctx, v.Client.B().Hgetall().Key(QUANTITY_MAP).Build()).AsStrMap()
	utils.HandleErrorWithLog(err)

	menuMapUnavailable := map[string]bool{}
	modifiedQuantityMap := map[string]int{}
	for key, _:= range(menuMap) {
		quantity, ok := quantityMap[key]
		if ok {	
			intQuantity, err := strconv.Atoi(quantity)
			utils.HandleErrorWithLog(err)
			modifiedQuantityMap[key] = intQuantity
			menuMapUnavailable[key] = true
			continue
		}
		menuMapUnavailable[key] = false
	}
	ofsbody := MenuMapOfs{ MenuMapAvailable: modifiedQuantityMap, MenuMapUnavailable: menuMapUnavailable}	
	err = JSONEncoderWrap(w, ofsbody)
	utils.HandleErrorWithLog(err)
}

func (v ValkeyModel) makeOfs(w http.ResponseWriter, r *http.Request) {
	fmt.Println("coming here")
	var menu MenuBody	
	err := JSONDecodeWrap(r.Body, &menu)
	utils.HandleErrorWithLog(err)
	quantityMap, err := v.Client.Do(v.Ctx, v.Client.B().Hgetall().Key(QUANTITY_MAP).Build()).AsStrMap()
	utils.HandleErrorWithLog(err)
	fmt.Printf("%v", menu)
	for _, ele := range menu.MenuList {
		key := ele.Itemcode
		stringStoredQantity, ok := quantityMap[key] 
		if !ok {
			stringStoredQantity = "0"
		}
		storedQuantity, err := strconv.Atoi(stringStoredQantity)
		quantity := storedQuantity - ele.Itemquantity		
		fmt.Printf("quantity %d", quantity)
		fmt.Printf("key %s", key)
		if quantity <= 0 {
			fmt.Println("coming here")
			err := v.Client.Do(v.Ctx, v.Client.B().Hdel().Key(QUANTITY_MAP).Field(key).Build()).Error()	
			if err != nil {
				utils.HandleErrorWithLog(err)
				break
			}
		}
		fmt.Println("coming here")
		err = v.Client.Do(v.Ctx, v.Client.B().Hset().Key(QUANTITY_MAP).FieldValue().FieldValue(key, strconv.Itoa(quantity)).Build()).Error()
		if err != nil {
			utils.HandleErrorWithLog(err)
			break
		}
	}
	err = JSONEncoderWrap(w, menu)
	utils.HandleErrorWithLog(err)
		
}
