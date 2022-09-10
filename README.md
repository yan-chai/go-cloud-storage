# Project 4: Cloud Storage

See project spec here: https://www.cs.usfca.edu/~mmalensek/cs521/assignments/project-4.html

You have quite a bit of freedom in this project, so be sure to write about your design here and explain how to use your system.

## 1. About This Project
   
* In this project, we will build a drop-box-like cloud storage service application. To operate, user needs to provide four argument inputs: 
     * ip address of the server user would like to connect, to make our service more safe and realiable, we have two servers, one is acting as main server, the other is replca, user can decide whcih server he/she wants to connect to. 
     * An action user wants to take against a file. There are four basic actions a user can use:
        *  put -- by calling put, the user can upload the mentioned file to the selected server
        *  get -- by caliing get, the user can download the mentioned file from the seleted server
        *  delete -- by calling put, the user can delete the mentioned file from both servers
        *  search -- by call search, user can get list of names of related files from server, if input file is blank, then all files' names from server directory will be returned, otherwise, names which contain the input file name string will be returned.
     

## 2. Logic And Implementation
* Either sending request from client to server or server to server, or receiving response from server to client or server to server, we need a "container" to hold such request or response information, to achieve this, we build a message "class". Message class contains the structure of a message, and method that can be implemented on message. There are two message struct we use for this project:
    * Message: which contains the following information: 
        1. user's request type, we use a number to represent it, 0 represents "put", 1 represents "get", 2 represent "search", 3 represents "delete". 
        2. filesize
        3. filename
        4. file's check sum result, we have a helper function called computeMd5 to do this, the return value of which is our check sum result. 
        5. backup addresss, we used addresses, one is localhost:9999, the other localhost:9998, and user can decide which one as the main, for example, if user's fourth argument is localhost:9999, the the main.go in client will automatically deicide localhost:9998 as the backup server's address.
    * MessageReply: which contains the following information:
        1. Success bool, which indicates if server implements the request action successfully.
	    2. Response string, which is a sentence telling client the status of the implementaion result, such as "operation success", "file does not exist" an so on.
	    3. FileSize int64, which is mainly used when user requests "get", so the server can tell client the size of the file user would like to get, then client can know in advance how big the file, also client can know what number it needs to put as the third argument of copyN function.
	    4. Files []string, which contain all file names that has the user's input filename as substring, for example, if server/backup server has two files, one is test1.txt, the other is test2.txt, and user's third argument is "t" or "test", this []string will be test1.txt test2.txt, since test1.txt and test2.txt have substing of "test" or "t".

* client's main function is responsible for reading commands from user and send corresponding request to required server, and then receive response sent from server, and then let user know his/her request result, if request in interuppted or fails, client would either let user resend the request or tell user such request can't be done.

* server's main function is responsible for receiving and dealing request sent by client, a server technically can deal requests sent by multiple clients by using multiple go routines. After sever receives request from client, before sending reply ,either indicating success or failure, server basically needs to do three things: 
    * First server will take action accordingly, for example, if server recives "put" command from client, server first needs to create a new file which has the same name received from user, information reagrding file name is sent via message "object", which can be retrived by checking message, then copy the file content to this newly created file using copyN function, by using copyN, server can know where to stop writing based on the third argument passed to copyN,which is the filesize, which can also be retrieved from message, after corresponding "put" implementation finishes, a new reply message would be built, which is used to store the information regarding the mplementaion result status, either success or failure.
    * Then server also needs to communicate with back-up server to request backserver take the same action, and then receives reply message from backup server.
    * Now server can send reply back to client based on implemention of both itself and backup server.

* backup's main function is responsible for receiving and dealing request sent by server, the implementation in it is pretty much the same as what we implement in server.
    
         
## 3.Demo

* A demo video is included in our github repo.

* This is what is shown in terminal during running backup: running backup:
```
[sjiang29@SJVM P4-go-hahahaha]$ ls 
 cloudStore  'Lab 7'   README.md
[sjiang29@SJVM P4-go-hahahaha]$ cd cloudStore/
[sjiang29@SJVM cloudStore]$ ls
backup  client  go.mod  message  server  utils
[sjiang29@SJVM cloudStore]$ cd backup/
[sjiang29@SJVM backup]$ go run main.go
stat test1.txt: no such file or directory
5a105e8b9d40e1329780d62ea2265d8a
stat test1.txt: no such file or directory
5a105e8b9d40e1329780d62ea2265d8a
```    
* This is what is shown in terminal during running server
```
[sjiang29@SJVM P4-go-hahahaha]$ ls 
 cloudStore  'Lab 7'   README.md
[sjiang29@SJVM P4-go-hahahaha]$ cd cloudStore/
[sjiang29@SJVM cloudStore]$ ls
backup  client  go.mod  message  server  utils
[sjiang29@SJVM cloudStore]$ cd server
[sjiang29@SJVM server]$ go run main.go
stat test1.txt: no such file or directory
5a105e8b9d40e1329780d62ea2265d8a
Operation Success! File successfully added to the sever
Operation Success! File successfully deleted on the sever.
stat test1.txt: no such file or directory
5a105e8b9d40e1329780d62ea2265d8a
Operation Success! File successfully added to the sever

```
* This is what is shown in terminal during running client
```
[sjiang29@SJVM P4-go-hahahaha]$ ls 
 cloudStore  'Lab 7'   README.md
[sjiang29@SJVM P4-go-hahahaha]$ cd cloudStore/
[sjiang29@SJVM cloudStore]$ ls
backup  client  go.mod  message  server  utils
[sjiang29@SJVM cloudStore]$ cd client/
[sjiang29@SJVM client]$ go run main.go put test1.txt localhost:9999
Operation Success! File successfully added to the sever.
[sjiang29@SJVM client]$ go run main.go delete test1.txt localhost:9999
Operation Success! File successfully deleted on the sever.
[sjiang29@SJVM client]$ go run main.go put test1.txt localhost:9999
Operation Success! File successfully added to the sever.
[sjiang29@SJVM client]$ go run main.go get test1.txt localhost:9999
Operation Success! File successfully retrieved
[sjiang29@SJVM client]$ go run main.go search test1.txt localhost:9999
Operation Success! Matching Files Found
test1.txt
```
