package main

import (
  "fmt"
  "net/http"
  "strings"
  "github.com/garyburd/redigo/redis"
  "regexp"
)
var p = fmt.Println


// =========================
type Todo struct {
  Con redis.Conn
  Data []string
}

func NewTodo() *Todo{
  todo := new(Todo)
  c, err := redis.Dial("tcp", "slackbot-redis:6379")
  if err != nil {
    panic(err)
  }
  //defer c.Close()
/*
  tasksString, err := redis.String(c.Do("GET", "todo"))
  if err != nil {
    tasksString = "[]";
  }
  err = json.Unmarshal(([]byte)(tasksString), &todo.Data)
  if err != nil {
    todo.Data = make([]string, 0)
  }
*/
  todo.Con = c
  return todo
}
func (todo *Todo)Close() {
  todo.Con.Close()
}
func (todo *Todo) add(message string) bool {
  _, err := todo.Con.Do("LPUSH", "todo", message) 
  if err != nil {
    p(err)
  }
  return true
}
func (todo *Todo) del(message string) bool {
  _, err := todo.Con.Do("LREM", "todo", 0, message) 
  if err != nil {
    p(err)
  }
 
  // ===============
  return true
}
/*
func (todo *Todo) delFromString(target string) bool {
  for _, v := range(todo.Data) {
  }
  // ===============
  tasksBytes, _ := json.Marshal(todo.Data)
  _, _ = todo.Con.Do("SET", "todo", tasksBytes) 
  // ===============
  return true
}
*/


func (todo *Todo) list() []string {
  tasksStrings, err := redis.Strings(todo.Con.Do("LRANGE", "todo", 0, -1))
  if err != nil {
    p(err)
  }
 
  return tasksStrings
}
// =========================


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

func sentence1(text string){
  regexp.MustCompile(``)
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
  todo := NewTodo()
  defer todo.Close()
  if todo.add(message) {
    return "追加しました"
  }
  return "何かおかしいです"
}
func del(message string) string {
  todo := NewTodo()
  defer todo.Close()

  if todo.del(message) {
    return "削除しました"
  }
  return "何かおかしいです"
}

func list() string {
  todo := NewTodo()
  defer todo.Close()

  ret := "todo:\n"
  for i, v := range todo.list() {
    ret += fmt.Sprintf("* [%d] %s\n", i, v)
  }
  return ret
}

func parseText(text string) (command string, post_text string) {
  command = getCommand(text)
  post_text = getMessage(text)
  return
}

func regMatch(text string,pattern string, index int) (string, bool) {
  add := regexp.MustCompile(pattern);
  sl := add.FindStringSubmatch(text)
  if len(sl) != 0 {
    return sl[index], true;
  } else {
    return "", false
  }
}

func process(text string) (string, bool) {
  var message string
  // =======================
  var s string
  var err bool
  if s, err = regMatch(text, `(.*)が(無い|ない)`, 1); err != false {
    message = add(s);
  } else if s, err = regMatch(text, `一覧`, 0); err != false {
    message = list();
  } else if s, err = regMatch(text, `(.*)(かった|買った|かいました|買いました)`, 1); err != false {
    message = del(s);
  }

  // =======================

/*
  if !validateParams(text) {
    return "unknown command", false
  }
  switch getCommand(text) {
    case "add": message = add(getMessage(text))
    case "del": message = del(getMessage(text))
    case "list": message = list()
  }
*/
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
