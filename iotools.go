package tools

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"
)

var ErrBadFmt = errors.New("bad format")
var ErrNoSuch = errors.New("no such")
var ErrDupData = errors.New("dup data")

func ReadConfig(fn string, conf interface{}) error {
	bt, err := ReadFile(fn)
	if err != nil {
		return err
	}
	return json.Unmarshal(bt, conf)
}

func RandInt(w int) int32 {
	rand.Seed(time.Now().UnixNano())
	return rand.Int31n(int32(w))
}

func RandStr(w int) string {
	rand.Seed(time.Now().UnixNano())
	base := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	str := ""
	for i := 0; i < w; i++ {
		idx := rand.Int31n(int32(len(base)))
		str = str + string(base[idx])
	}
	return str
}

func ReadFile(fn string) ([]byte, error) {
	file, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return ioutil.ReadAll(file)
}

func DoPost(url string, v *url.Values) ([]byte, error) {
	resp, err := http.PostForm(url, *v)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func DoGet(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

type LineFunc func(line string) error

func GetAsLine(uri string, lf LineFunc) error {
	bt, err := DoGet(uri)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(bt)
	for {
		line, err := buf.ReadString('\n')
		if err != nil && err != io.EOF {
			return err
		}
		if err == io.EOF {
			break
		}
		if line == "" {
			continue
		}
		line = line[:len(line)-1]
		err = lf(line)
		if err != nil {
			return err
		}
	}
	return nil
}

func ReadLineArr(fn string, lf LineFunc, split int) error {
	return readLine(fn, lf, split)
}

func ReadLine(fn string, lf LineFunc) error {
	return readLine(fn, lf, 0)
}

func readLine(fn string, lf LineFunc, split int) error {
	file, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer file.Close()
	rd := bufio.NewReader(file)
	for {
		line, err := rd.ReadString('\n')
		if err != nil && err != io.EOF {
			return err
		}
		if err == io.EOF {
			break
		}
		if line == "" {
			continue
		}
		line = line[:len(line)-1]
		if split != 0 {
			arr := strings.Split(line, " ")
			if len(arr) != split {
				return ErrBadFmt
			}
		}
		err = lf(line)
		if err != nil {
			return err
		}
	}
	return nil
}

func Proxy(c *net.TCPConn) {
	defer c.Close()
	now := time.Now()
	fmt.Println(now, "we get an conn from", c.RemoteAddr())
	fmt.Println(now, "and we are going to 119.28.77.61:8000...")
	var raddr net.TCPAddr
	raddr.IP = net.ParseIP("119.28.77.61")
	raddr.Port = 8000
	r, err := net.DialTCP("tcp4", nil, &raddr)
	if err != nil {
		fmt.Println("dial remote", err)
		return
	}
	go io.Copy(c, r)
	io.Copy(r, c)
}

func Run() error {
	var addr net.TCPAddr
	addr.Port = 8080
	ls, err := net.ListenTCP("tcp4", &addr)
	if err != nil {
		return err
	}
	for {
		c, err := ls.AcceptTCP()
		if err != nil {
			fmt.Println("accept error", err)
			continue
		}
		go Proxy(c)
	}
}

func BCrypt(pass string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
}

func BDecrypt(hash []byte, pass string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass))
	if err != nil {
		return false
	}
	return true
}

// a joke
func fan(ch chan int) {
	i := uint64(1)
	for {
		i = i + 1
	}
}

func StartFan() {
	ch := make(chan int)
	ncpu := runtime.NumCPU()
	if runtime.GOMAXPROCS(ncpu) < 0 {
		panic("could not set maxprocs")
	}
	for i := 0; i < ncpu; i++ {
		go fan(ch)
	}
	<-ch
}

func FileMd5(fn string) (int, string, error) {
	file, err := os.Open(fn)
	if err != nil {
		return 0, "", err
	}
	defer file.Close()
	bt, err := ioutil.ReadAll(file)
	if err != nil {
		return 0, "", err
	}
	return len(bt), fmt.Sprintf("%x", md5.Sum(bt)), nil
}

func UnixToBJ(unix int64) time.Time {
	return unixTo(unix, "Asia/Shanghai")
}

func UnixToUSPacific(unix int64) time.Time {
	return unixTo(unix, "US/Pacific")
}

func UnixToUTC(unix int64) time.Time {
	return unixTo(unix, "UTC")
}

func unixTo(unix int64, name string) time.Time {
	l, err := time.LoadLocation(name)
	if err != nil {
		panic("bad time name : " + name)
	}
	t := time.Unix(unix, 0)
	return t.In(l)
}

func TimeStr() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func FormatTime(sec int64) string {
	t := time.Unix(sec, 0)
	return t.Local().Format("2006-01-02 15:04:05")
}
