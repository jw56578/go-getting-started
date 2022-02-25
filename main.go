package main

import (
  "bytes"
  "io/ioutil"
  "log"
  "net/http"
  "fmt" 
  "encoding/json"
  "strconv"
   "net/mail"
  //"time"
)

var recipientState = make(map[int]string)
var recipientToContainer = make(map[int]int)

 /******TODO



Questions:
1. what is P2 va PUSHPULL
2. the message_reply call back doesn't have a container id, ask rafeal what is the best way to handle this


download file from push pull
https://botdoc.atlassian.net/wiki/spaces/BOTDOC/pages/287146013/API+How+to+download+your+files+from+a+P2+-+Pull


// dashboard
// https://sandboxdev.botdoc.io/dashboard
// -- when you manually do the process in the dashboard, the requests will be shown at the end in curl
// make sure you know where the api docs are
// -- https://api.botdoc.io/documentation/#
// https://botdoc.atlassian.net/wiki/spaces/BOTDOC/pages/39485516/API+Botdoc+Postman+Collection
  // --- download postman collection to understand the workflow


*/////

 /// repl.it is not working to run multiple files. You have to do it manually from the shell
 //go run main.go get_token.go

func getToken() string{
  postBody, _ := json.Marshal(map[string]string{
    "api_key":  "e700385ee663e8e4f8398f2bc36707bc4b93f181",
    "email": "jon.woo@cdk.com",
  })
  responseBody := bytes.NewBuffer(postBody)
  //Leverage Go's HTTP Post function to make request
  resp, err := http.Post("https://sandboxapi.botdoc.io/v1/auth/get_token/", "application/json", responseBody)
  //Handle Error
  if err != nil {
    log.Fatalf("An Error Occured %v", err)
  }
  defer resp.Body.Close()
  //Read the response body
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    log.Fatalln(err)
  }
  sb := string(body)
  data := TokenResponse{}
  json.Unmarshal([]byte(sb), &data) 
  return data.Token
}

func main() {
    // the servemux, in exactly the same way that we did before.
    mux := http.NewServeMux()

    mux.HandleFunc("/", HelloHandler)
    //mux.HandleFunc("/email", EmailHandler) <--- this doesn't work, does it need to be in order?

    fmt.Println("Server started at port 8080")
    log.Fatal(http.ListenAndServe(":8080", mux))
}
func EmailHandler(w http.ResponseWriter, r *http.Request) {
    keys, ok := r.URL.Query()["email"]

  


    addr, err := mail.ParseAddress(keys[0])
    if err != nil {
         fmt.Fprintf(w, 
                      `<!DOCTYPE html>
            <html lang="en">
            <head>
                <meta charset="UTF-8">
                <title>About page</title>
            </head>
            <body>
              Invalid email %s
            </body>
            </html>`, keys[0])
      return
    }

  
    if ok {
         fmt.Fprintf(w, 
                      `<!DOCTYPE html>
            <html lang="en">
            <head>
                <meta charset="UTF-8">
                <title>About page</title>
            </head>
            <body>
               A container has been created for %s!
                <br/>
               Please check email for link
            </body>
            </html>`, addr.Address)
      runWorkflow(addr.Address)
    }
}
func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	var bodyBytes []byte
	var err error

	if r.Body != nil {
		bodyBytes, err = ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Printf("Body reading error: %v", err)
			return
		}
		defer r.Body.Close()
	}

	//fmt.Printf("Headers: %+v\n", r.Header)

	if len(bodyBytes) > 0 {
    data := BotDocCallback{}
    json.Unmarshal(bodyBytes, &data) 
    fmt.Print("\n")
    fmt.Print("The type of callback recieved is: " + data.Type)
    if(data.Type == "message_reply"){
      var prettyJSON bytes.Buffer
  		if err = json.Indent(&prettyJSON, bodyBytes, "", "\t"); err != nil {
  			fmt.Printf("JSON parse error: %v", err)
  			return
  		}
  		fmt.Println(string(prettyJSON.Bytes()))
      responseMessageReply(data.MessageReply.Recipient)
    }
    if(data.Type == "feature" && data.Feature.State == "complete"){
      var prettyJSON bytes.Buffer
  		if err = json.Indent(&prettyJSON, bodyBytes, "", "\t"); err != nil {
  			fmt.Printf("JSON parse error: %v", err)
  			return
  		}
  		fmt.Println(string(prettyJSON.Bytes()))
      // WILL THIS HAVE THE PULL ID
      handleFileUploaded(data.Feature.Container, data.Feature.Pull.Id)
    }


	} else {
		fmt.Printf("Body: No Body Supplied\n")
	}
}
func HelloHandler(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path == "/media" {
         MediaHandler(w,r)
    }
    if r.URL.Path == "/email" {
         EmailHandler(w,r)
    }
     if r.URL.Path == "/callback" {
         CallbackHandler(w,r)
    }
     if r.URL.Path == "/" {
           fmt.Fprintf(w, `
            <!DOCTYPE html>
            <html lang="en">
            <head>
                <meta charset="UTF-8">
                <title>About page</title>
            </head>
            <body>
                <script>
                  var email = prompt('Enter Email');
                  location.href = '/email?email=' + email;
        
                </script>
            </body>
            </html>       
            `)
    }

 
}
func MediaHandler(w http.ResponseWriter, r *http.Request) {
  
   keys, ok := r.URL.Query()["id"]
   fmt.Println("the pull id recieved was  " + keys[0])
    if ok {
      pullId, err := strconv.Atoi(keys[0])
      if(err != nil){
        
      }
        var token = getToken()
        var fileBytes = getMedia(token, pullId)
      	w.WriteHeader(http.StatusOK)
      	w.Header().Set("Content-Type", "application/octet-stream")
      	w.Write(fileBytes)

    }
/*
  fileBytes, err := ioutil.ReadFile("test.png")
	if err != nil {
		       fmt.Fprintf(w, `
          <!DOCTYPE html>
          <html lang="en">
          <head>
              <meta charset="UTF-8">
              <title>About page</title>
          </head>
          <body>
              Error, trying to find the media
          </body>
          </html>       
          `)
	}
*/

	return

  
}
func handleFileUploaded(cId int, pullId int){
 fmt.Print("sending a message to the user with the link to the file")
  //recipientState[recipientId] = "waitingforinsurance"
  var token = getToken()
  // send a message that has the link to view the file
  createMessage(token,cId,"Link to file.  https://botdocpoc.jw56578.repl.co/media?id=" + strconv.Itoa(pullId))


  
}
func responseMessageReply(recipientId int){
  var token = getToken()
  fmt.Print(token)


  // hard coded that they replied with their first name, hard code asking for last name next 
  // THIS WORKS AS IS, NO NEED TO CALL sendNotification
  /*******
  THE message_reply hook does not contain the container ID
  how do you get this???
  we could relate it to the recipient id
  *******///
  fmt.Print("\n")
  fmt.Print("\n")
  if(recipientState[recipientId] == "askforlastname"){
    fmt.Println("asking for last name")
    createMessage(token,recipientToContainer[recipientId], "Thank you. Please provide your last name.")
    recipientState[recipientId] = "askfordriverslicense"
  } else if(recipientState[recipientId] == "askfordriverslicense"){
    createPull(token,recipientToContainer[recipientId])
    recipientState[recipientId] = "waitingfordriverslicense"
  }
  
}
func runWorkflow(email string) {
 //ExternalThing()

  var token = getToken()
  //createRequest(token)
  var cId = createContainer(token)
  var rId = createRecipient(token, cId)
  
  recipientToContainer[rId] = cId
  recipientState[rId] = "askforlastname"
  
  createRecipientMethods(token,rId, email)
  createEmail(token,cId)

  fmt.Print("\n")
  fmt.Print("\n")
  // this is creating a random dialog box that says, Pending and Sign, I have no idea what that means
  //createPullFeature(token,cId)

  // this needs to be done first to ask for first name and last name
  fmt.Print("\n")
  fmt.Print("\n")
  createMessage(token,cId, "Hello. Please provide your first name.")

  
  fmt.Print("\n")
  fmt.Print("\n")
  sendNotification(token,cId)
}

// this does not cause any email to be sent, even though interface_class is set to email
func createRequest(token string){

   var jsonData = []byte(`{
      "message": "Hello",
      "requester_privatenotes": "",
      "type": "push",
      "is_draft": true,
      "long_message_subject": "DL required",
      "short_message": "Please upload DL",
      "callback_url": "https://myfirstwebserver.jw56578.repl.co",
      "long_message": "<p>Hi John,</p><p><br></p><p>Please upload DL.</p>",
      "contact": [
          {
              "first_name": "Sheni",
              "last_name": "Singhal",
              "contactmethod": [
                  {
                      "interface_class": "email",
                      "value": "jw56578@gmail.com"
                  }
              ]
          }
      ]
  }`)
	request, err := http.NewRequest("POST", "https://sandboxapi.botdoc.io/v1/request/" , bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
  request.Header.Set("Authorization", "JWT " + token)
  if err != nil {
    log.Fatalf("An Error Occured %v", err)
  }

  client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	// b, err := ioutil.ReadAll(resp.Body)  Go.1.15 and earlier
	if err != nil {
		log.Fatalln(err)
	}

  fmt.Printf("The response of creating a request is: %s", b)



/*
  defer request.Body.Close()
  //Read the response body
  body, err := ioutil.ReadAll(request.Body)
  if err != nil {
    log.Fatalln(err)
  }
  sb := string(body)
  fmt.Printf("The response of creating a request is: %s", sb)
    */
}

// what is create request VS this other workflow?????
func createContainer(token string) int{

   var jsonData = []byte(`{
    "callback_url": "https://BotDocPOC.jw56578.repl.co/callback",
    "page_type": "p2",
    "display_chat": true
  }`)
	request, err := http.NewRequest("POST", "https://sandboxapi.botdoc.io/v1/module_container/container/" , bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
  request.Header.Set("Authorization", "JWT " + token)
  if err != nil {
    log.Fatalf("An Error Occured %v", err)
  }

  client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

  //fmt.Printf("The response of creating a container is: %s", b)

  data := CreateContainerResponse{}
  json.Unmarshal([]byte(b), &data) 
  fmt.Print("Container Id: " + strconv.Itoa(data.ID))
  return data.ID

}
func createRecipient(token string, containerId int) int{

   var jsonData = []byte(`{
    "first_name": "New",
    "last_name": "Customer"
  }`)
	request, err := http.NewRequest("POST", "https://sandboxapi.botdoc.io/v1/module_container/container/" + strconv.Itoa(containerId) + "/recipient/" , bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
  request.Header.Set("Authorization", "JWT " + token)
  if err != nil {
    log.Fatalf("An Error Occured %v", err)
  }

  client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		  fmt.Printf("error is: %s", err)
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	// b, err := ioutil.ReadAll(resp.Body)  Go.1.15 and earlier
	if err != nil {
		 fmt.Printf("error is: %s", err)
	}

  //fmt.Printf("The response of creating a recipient is: %s", b)

  data := CreateRecipientResponse{}
  json.Unmarshal([]byte(b), &data) 
 
  return data.ID
}
func createRecipientMethods(token string, recipientId int, email string) int{
  
  jsonData, _ := json.Marshal(map[string]string{
    "interface_class":  "email",
    "value": email,
    //can add recepient id
  })

	request, err := http.NewRequest("POST", "https://sandboxapi.botdoc.io/v1/module_container/recipient/" + strconv.Itoa(recipientId) + "/recipient_item/" , bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
  request.Header.Set("Authorization", "JWT " + token)
  if err != nil {
    log.Fatalf("An Error Occured %v", err)
  }

  client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		  fmt.Printf("error is: %s", err)
	}

	defer resp.Body.Close()

  //fmt.Printf("The response of creating a recipient method is: %s", b)

  return 0
}
func createEmail(token string, containerId int) int{

   var jsonData = []byte(`{
    "subject": "A new document has been requested by a dealer",
    "body": "Please click on this link and follow the instruction to get your test drive started"
  }`)
	request, err := http.NewRequest("POST", "https://sandboxapi.botdoc.io/v1/module_container/container/" + strconv.Itoa(containerId) + "/email/" , bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
  request.Header.Set("Authorization", "JWT " + token)
  if err != nil {
    log.Fatalf("An Error Occured %v", err)
  }

  client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		  fmt.Printf("error is: %s", err)
	}

	defer resp.Body.Close()

  //fmt.Printf("The response of creating a an email subject, body is: %s", b)

  return 0
}

func createPush(token string, containerId int) int{

  jsonData, _ := json.Marshal(map[string]string{
    "message":  "does this go into the chat message?",
    "type": "push",
    "container":strconv.Itoa(containerId),
  })

  
	request, err := http.NewRequest("POST", "https://sandboxapi.botdoc.io/v1/module_container_pushpull/pushpullfeature/", bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
  request.Header.Set("Authorization", "JWT " + token)
  if err != nil {
    log.Fatalf("An Error Occured %v", err)
  }

  client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		  fmt.Printf("error is: %s", err)
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	// b, err := ioutil.ReadAll(resp.Body)  Go.1.15 and earlier
	if err != nil {
		 fmt.Printf("error is: %s", err)
	}

  fmt.Printf("The response of creating a PUSH: %s", b)
  return 0
}


func createPullFeature(token string, containerId int) int{

  jsonData, _ := json.Marshal(map[string]string{
    "message":  "where does the message key/value of creating a pull go?",
    "type": "pull",
    "container":strconv.Itoa(containerId),
  })

  
	request, err := http.NewRequest("POST", "https://sandboxapi.botdoc.io/v1/module_container_pushpull/pushpullfeature/", bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
  request.Header.Set("Authorization", "JWT " + token)
  if err != nil {
    log.Fatalf("An Error Occured %v", err)
  }

  client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		  fmt.Printf("error is: %s", err)
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	// b, err := ioutil.ReadAll(resp.Body)  Go.1.15 and earlier
	if err != nil {
		 fmt.Printf("error is: %s", err)
	}

  fmt.Printf("The response of creating PULL Feature: %s", b)

  //data := CreateRecipientResponse{}
  //json.Unmarshal([]byte(b), &data) 
  //fmt.Print("Container Id: " + strconv.Itoa(data.ID))
  //return data.ID
  return 0
}




func createPull(token string, containerId int) int{

  jsonData, _ := json.Marshal(map[string]string{
    "description":  "Please provide your license so we know you are legally allowed to drive",
    "title": "Drivers License Upload",
    "container":strconv.Itoa(containerId),
  })

  
	request, err := http.NewRequest("POST", "https://sandboxapi.botdoc.io/v1/module_container_pull/pull/", bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
  request.Header.Set("Authorization", "JWT " + token)
  if err != nil {
    log.Fatalf("An Error Occured %v", err)
  }

  client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		  fmt.Printf("error is: %s", err)
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	// b, err := ioutil.ReadAll(resp.Body)  Go.1.15 and earlier
	if err != nil {
		 fmt.Printf("error is: %s", err)
	}

  fmt.Printf("The response of creating PULL: %s", b)

  //data := CreateRecipientResponse{}
  //json.Unmarshal([]byte(b), &data) 
  //fmt.Print("Container Id: " + strconv.Itoa(data.ID))
  //return data.ID
  return 0
}


func createMessage(token string, containerId int ,message string) int{

  jsonData, _ := json.Marshal(map[string]string{
    "body": message,
    "container":strconv.Itoa(containerId),
  })

  
	request, err := http.NewRequest("POST", "https://sandboxapi.botdoc.io/v1/module_container/message/", bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
  request.Header.Set("Authorization", "JWT " + token)
  if err != nil {
    log.Fatalf("An Error Occured %v", err)
  }

  client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		  fmt.Printf("error is: %s", err)
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	// b, err := ioutil.ReadAll(resp.Body)  Go.1.15 and earlier
	if err != nil {
		 fmt.Printf("error is: %s", err)
	}

  fmt.Printf("The response of creating a message: %s", b)

  //data := CreateRecipientResponse{}
  //json.Unmarshal([]byte(b), &data) 
  //fmt.Print("Container Id: " + strconv.Itoa(data.ID))
  //return data.ID
  return 0
}



func getMedia(token string, pullId int) []byte {

  var url = "https://sandboxapi.botdoc.io/v1/module_container_pull/pull/" + strconv.Itoa(pullId) + "/pull_file/"
  fmt.Printf("getting media for a container url is: %s", url)
	request, err := http.NewRequest("GET",url ,nil)
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
  request.Header.Set("Authorization", "JWT " + token)
  if err != nil {
    log.Fatalf("An Error Occured %v", err)
  }

  client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		  fmt.Printf("error is: %s", err)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	// b, err := ioutil.ReadAll(resp.Body)  Go.1.15 and earlier
	if err != nil {
		 fmt.Printf("error is: %s", err)
	}
  fmt.Printf("The response of getting the media for a container is: %s", b)

  var arr []PullFileResponse
  json.Unmarshal([]byte(b), &arr)
  log.Printf("Unmarshaled: %v", arr[0].DownloadUrl)
  return downloadFile(token, arr[0].DownloadUrl); 
  
}

func downloadFile(token string, url string) []byte{

  fmt.Printf("downloading this url: %s", url)
	request, err := http.NewRequest("GET",url ,nil)
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
  request.Header.Set("Authorization", "JWT " + token)
  if err != nil {
    log.Fatalf("An Error Occured %v", err)
  }

  client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		  fmt.Printf("error is: %s", err)
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		 fmt.Printf("error is: %s", err)
	}
 //time.Sleep(8 * time.Second)
  
  //fmt.Printf("The response of downloading a file is: %s", b)


  return b
}


// this is the function that will trigger the sending of an email to the customer
func sendNotification(token string, containerId int) int{

	request, err := http.NewRequest("GET", "https://sandboxapi.botdoc.io/v1/module_container/container/" + strconv.Itoa(containerId) + "/send_notification/",nil)
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
  request.Header.Set("Authorization", "JWT " + token)
  if err != nil {
    log.Fatalf("An Error Occured %v", err)
  }

  client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		  fmt.Printf("error is: %s", err)
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	// b, err := ioutil.ReadAll(resp.Body)  Go.1.15 and earlier
	if err != nil {
		 fmt.Printf("error is: %s", err)
	}

  fmt.Printf("The response of sending a notification: %s", b)

  //data := CreateRecipientResponse{}
  //json.Unmarshal([]byte(b), &data) 
  //fmt.Print("Container Id: " + strconv.Itoa(data.ID))
  //return data.ID
  return 0
}
/*







#HttpRequest made to send the container to the receivers
curl --request GET --header 'Content-Type: application/json' --header 'Accept: application/json' --header 'Authorization: JWT <your_jwt_token>' 'https://api.botdoc.io/v1/module_container/container/35811/send_notification/'

*/



// struct = object
// this is how you define a type to parse from json
type TokenResponse struct {
  Token string      `json:"token"`
}
type CreateContainerResponse struct {
  ID int      `json:"id"`
  Identifier string      `json:"identifier"`
}
type CreateRecipientResponse struct {
  ID int      `json:"id"`
}
type BotDocCallback struct {
  Type string      `json:"type"`
  //contact_notification_send
  //container  <------- container created
  //session_opened
  //message_reply
  //feature <----THIS IS WHAT HAPPENED WHEN USER UPLOADS FILE
  Container CreateContainerResponse      `json:"container"`
  MessageReply MessageReply      `json:"message_reply"`
  Feature Feature      `json:"feature"`
}

type MessageReply struct {
  Recipient int      `json:"recipient"`
}
type Feature struct {
  Id int      `json:"id"`
  Container int      `json:"container"`
  Title string      `json:"title"`
  State string      `json:"state"`
  Pull PullFileResponse      `json:"pull"`
}
type PullFileResponse struct {
  Id int      `json:"id"`
  ContentType string      `json:"content_type"`
  Extension string      `json:"extension"`
  DownloadUrl string      `json:"download_url"`
}


/*
{
    "type": "feature",
    "feature": {
        "id": 40577,
        "container": 36240,
        "created": "2022-02-25T02:07:26.392342Z",
        "updated": "2022-02-25T02:07:26.477922Z",
        "feature_attr": "pull",
        "title": "Drivers License Upload",
        "description": "Please provide your license so we know you are legally allowed to drive",
        "state": "complete",
        "expiration": "2022-02-28T02:06:18.962745Z",
        "custom_label_pending": null,
        "pull": {
            "id": 40577,
            "container": 36240,
            "allowed_extensions": {},
            "title": "Drivers License Upload",
            "description": "Please provide your license so we know you are legally allowed to drive",
            "state": "complete",
            "expiration": "2022-02-28T02:06:18.962745Z",
            "custom_label_pending": null,
            "custom_label_complete": null,
            "custom_label_expired": null,
            "created": "2022-02-25T02:07:26.392342Z",
            "updated": "2022-02-25T02:07:26.477922Z",
            "file_size_limit": null,
            "max_files": null,
            "recipient_permissions": []
        }
    },
    "callback_identifier": "f2e406577191dcab66813cf201bae4c1"
}


*/
