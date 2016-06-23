package main

import (
  "fmt"
  "net/http"
  "strings"
  "github.com/garyburd/redigo/redis"
  "encoding/json"
  "strconv"
)
var p = fmt.Println


func contains (str string, list []string) bool {
  for _, v := range(list) {
    if v == str {
      return true
    }
  }
  return false
}

func getTriggerWord(text string) string {
  return strings.Split(text, " ")[0]
}
func getCommand(text string) string {
  return strings.Split(text, " ")[1]
}
func getMessage(text string) string {
  return strings.Split(text, " ")[2]
}
func validateParams(text string) bool {
  if len(strings.Split(text, " ")) < 2 {
    return false
  }
  if getTriggerWord(text) != "todo" {
    return false
  }
  command := getCommand(text);
  correctCommands := []string{"del", "add","list"}
  if !contains(command, correctCommands){
    return false
  }
  return true
}


func add(message string) string {
  // ===============
  c, err := redis.Dial("tcp", "slackbot-redis:6379")
  if err != nil {
    panic(err)
  }
  defer c.Close()

  tasksString, err := redis.String(c.Do("GET", "todo"))
  if err != nil {
    tasksString = "[]";
  }
  var tasks []string
  err = json.Unmarshal(([]byte)(tasksString), &tasks)
  if err != nil {
    tasks = make([]string, 0)
  }
  // ===============
  tasks = append(tasks, message)

  p(tasks)
  // ===============
  tasksBytes, err := json.Marshal(tasks)
  p(tasksBytes)
  _, err = c.Do("SET", "todo", tasksBytes) 
  p(err) //nil
  // ===============

  return "追加しました"
}
func del(message string) string {
  // ===============
  c, err := redis.Dial("tcp", "slackbot-redis:6379")
  if err != nil {
    panic(err)
  }
  defer c.Close()

  tasksString, err := redis.String(c.Do("GET", "todo"))
  if err != nil {
    p("get error")
    tasksString = "[]";
  }
  var tasks []string
  err = json.Unmarshal(([]byte)(tasksString), &tasks)
  if err != nil {
    p("error unmarshal")
    tasks = make([]string, 0)
  }
  // ===============

  n, err := strconv.Atoi(message)
  if err != nil {
    return "引数のエラーです"
  }
  if n  >= len(tasks) {
    return "引数のエラーです"
  }
  tasks = append(tasks[:n], tasks[n+1:]...)

  p(tasks)
  // ===============
  tasksBytes, err := json.Marshal(tasks)
  p(tasksBytes)
  _, err = c.Do("SET", "todo", tasksBytes) 
  p(err) //nil
  // ===============

  return "削除しました"
}

func list() string {
  // ===============
  c, err := redis.Dial("tcp", "slackbot-redis:6379")
  if err != nil {
    panic(err)
  }
  defer c.Close()

  tasksString, err := redis.String(c.Do("GET", "todo"))
  if err != nil {
    p("get error")
    tasksString = "[]";
  }
  var tasks []string
  err = json.Unmarshal(([]byte)(tasksString), &tasks)
  if err != nil {
    p("error unmarshal")
    tasks = make([]string, 0)
  }
  // ===============

  ret := "todo:\n"
  for i, v := range tasks {
    ret += fmt.Sprintf("- %d %s\n", i, v)
  }
  return ret
}




func parseText(text string) (command string, post_text string) {
  command = getCommand(text)
  post_text = getMessage(text)
  return
}

func process(text string) (string, bool) {
  var message string

  if !validateParams(text) {
    return "unknown command", false
  }
  switch getCommand(text) {
    case "add": message = add(getMessage(text))
    case "del": message = del(getMessage(text))
    case "list": message = list()
  }
  return message, true
}



func todoListBot(w http.ResponseWriter, r *http.Request) {
  checkUser(w, r, func(text string, channel_name string) {
      returnText, err := process(text)
      //return_text := "'" + text + "'"
      p(err)
      if err == true {
        fmt.Fprintf(w, "{\"text\": \"%s\"}", returnText)
      }
  })
}

func checkUser(w http.ResponseWriter, r *http.Request, proc func(text string, channel_name string)) {
  if r.Method == "POST" {
      text := r.FormValue("text")
      user_name := r.FormValue("user_name")
      channel_name := r.FormValue("channel_name")

      if user_name != "slackbot" {
          p("user_name:", user_name)
          p("channel_name:", channel_name)
          proc(text, channel_name)
      }
  }}

func testHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello")
}
func main() {
  p("start server")
  http.HandleFunc("/todo", todoListBot)
  http.HandleFunc("/test", testHandler)
  http.ListenAndServe(":8888", nil)
}
