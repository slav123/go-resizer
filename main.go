package main

// http://stackoverflow.com/questions/11354518/golang-application-auto-build-versioning

import (
	"fmt"
	"resizer/fs"
	"resizer/helper"
	//"github.com/daddye/vips"
	"encoding/json"
	"flag"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"resizer/apachelog"
	"resizer/image"
	"resizer/rwrap"
	"resizer/s3"
	//	"resizer/sdb"
	"runtime"

	"strconv"
	"strings"
	"time"
	//	"resizer/sdb"
)

var gConfig = map[string]string{
	"port": "8080",
}

// determine max no of workers
var workers = runtime.NumCPU()

// src Dir
var srcDir, lastError string

// params
//var params map[string]string

// dirs
const (
	cacheDir = "cache/"
	version  = "2017030101"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	var configFile string

	// parse command line arguments
	flag.StringVar(&configFile, "c", "config.json", "Path to the configuration file")
	flag.Parse()

	// load in the config.
	b, err := storage.ReadFile(configFile)
	if err != nil {
		fmt.Println("Could not read config file: " + configFile)
	} else {
		fmt.Println("Reading config " + configFile)
	}

	var myConfig map[string]string

	err = json.Unmarshal(b, &myConfig)

	if err != nil {
		fmt.Println("Could not parse config file")
	}
	for k, v := range myConfig {
		gConfig[k] = v
	}

	fmt.Println("version: " + string(version))
}

func configure() {
	rwrap.Port = gConfig["redis_port"]
	rwrap.Server = gConfig["redis_serv"]
	rwrap.Start()

	//	sdb.AccessKeyId = gConfig["sdb.AccessKeyId"]
	//	sdb.SecretAccessKey = gConfig["sdb.SecretAccessKey"]
	//	sdb.Region = gConfig["sdb.Region"]
}

func main() {

	configure()

	fmt.Println("running at: " + gConfig["port"])

	r := mux.NewRouter()
	r.HandleFunc(`/status`, handleStatus)

	r.HandleFunc(`/{user}/{origin}/{action}/{object}`, UserOriginHandler)
	r.HandleFunc(`/{user}/{origin}/{action}/{object:[a-zA-Z0-9=\-/_.]+}`, UserOriginHandler)

	r.HandleFunc("/", hello)

	loggingHandler := apachelog.NewApacheLoggingHandler(r, os.Stdout)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", gConfig["port"]),
		Handler:      loggingHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Fatal(server.ListenAndServe())

}

func UserOriginHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	//user := vars["user"]
	origin := vars["origin"]

	cr := rwrap.Exists("origin:" + origin)
	if cr == false {
		//http.NotFound(w, r)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Origin :" + origin + ": doesn't exists "))
	}

	params := rwrap.MGetH("origin:" + origin)

	switch params["type"] {
	case "s3":
		s3.Bucket = params["bucket"]
		s3.AccessKeyId = params["accessKeyId"]
		s3.SecretAccessKey = params["secretAccessKey"]
		s3.Region = params["region"]
		s3.Expire, _ = strconv.Atoi(params["expire"])
		rwrap.Expire, _ = strconv.Atoi(params["expire"])
		srcDir = ""
		handleResizeS3(w, r, params)
	case "ws":
		s3.Bucket = params["bucket"]
		s3.AccessKeyId = params["accessKeyId"]
		s3.SecretAccessKey = params["secretAccessKey"]
		s3.Region = params["region"]
		srcDir = params["url"]
		handleResizeS3(w, r, params)
	case "fs":
		srcDir = params["src"]
		handleResizeFs(w, r, params)
	default:

	}

}

// handle request to resize FS
func handleResizeFs(w http.ResponseWriter, r *http.Request, params map[string]string) {

	var (
		ext    string
		err    error
		res    []byte
		re     []byte
		op     helper.ImageOp
		cached bool
	)

	// prepare hashes and data
	op = helper.Prepare(r.URL.Path)

	// ad image to owner
	vars := mux.Vars(r)
	rwrap.SADD(vars["user"]+":"+"objects", op.Hash)

	// get flush, and ignore cache
	flush := r.URL.Query().Get("flush")

	if flush == "1" {
		cached = false
	} else {
		// check if file exists
		cached = storage.CheckCacheFs(op.Hash, cacheDir)
	}

	// get correct source file
	if cached {
		res, err = storage.ReadFile(cacheDir + op.Hash)
	} else {
		// read file
		re, err = storage.ReadFile(srcDir + op.Filename)
	}

	// if we failed to read return message
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Failed to read source file " + srcDir + op.Filename))
	} else {

		if strings.Contains(r.UserAgent(), "Chrome") && op.Format == 0 && params["convert"] == "on" {
			op.Format = 2
		}

		t := time.Now().Local()

		if !cached {
			res = image.Resize(re, op)
			ext = image.DetermineImageTypeName(&res)
			//	ext = strings.ToLower(op.Filename[strings.LastIndex(op.Filename, ".")+1:])
			log.Println("new request :" + op.Filename)
			log.Println("file type:" + ext)
			go storage.SaveFile(res, cacheDir+op.Hash)
		}
		//ext = strings.ToLower(op.Filename[strings.LastIndex(op.Filename, ".")+1:])

		ext = image.DetermineImageTypeName(&res)
		log.Printf("after conversion %s, type %d", ext, strconv.Itoa(len(res)))

		w.Header().Set("Content-Type", helper.GetImageType(ext))
		w.Header().Set("Content-Length", strconv.Itoa(len(res)))
		w.Header().Set("Date", t.Format(time.RFC1123))
		w.Write(res)

	}
}

// handle request to resize
func handleResizeS3(w http.ResponseWriter, r *http.Request, params map[string]string) {

	var (
		ext, etag, lmod string
		err             error
		res, re         []byte
		cached, ok      bool
		size            int
	)

	// prepare
	op := helper.Prepare(r.URL.Path)

	// assign object to user
	vars := mux.Vars(r)
	rwrap.SADD(vars["user"]+":"+"objects", op.Hash)

	// get referrer if exists
	ref := r.Referer()
	if ref != "" {
		log.Println("refferer: " + ref)
	}

	// look for redirect field
	/*
		redirect := r.URL.Query().Get("r")

		if redirect == "1" {
			fname := cache.GetH(op.Hash, "fname")
			// https://s3-ap-southeast-2.amazonaws.com/photo.coreadvisory.com.au/cache/317c44af8753b89ce8072815f23cfadf.jpeg
			// 2DO: get corrrect bucket from cache
			http.Redirect(w, r, "http://s3-ap-southeast-2.amazonaws.com/"+gConfig["bucket"]+"/"+fname, 301)

			return
		}*/

	// get flush command
	flush := r.URL.Query().Get("flush")

	// create hash name
	s := []string{"ob:", vars["user"], ":", vars["origin"], ":", op.Hash}

	// create key Object
	keyObject := strings.Join(s, "")

	if flush == "1" {

		go purge("http://rsize.xyz/" + r.URL.Path)
		rwrap.Del(keyObject)
		cached = false
	} else {
		cached = rwrap.Exists(keyObject)
	}

	t := time.Now().Local()

	// get source file
	if cached {
		meta := rwrap.MGetH(keyObject)

		fname := meta["fname"]
		etag = meta["md5"]

		if lmod, ok = meta["date"]; !ok {
			lmod = t.Format(time.RFC1123)
		}

		//log.Printf("Got cached copy %+v", meta)

		res, err = s3.Get(fname)

		size = len(res)

		if size == 0 {
			log.Println("Failed to pull cached copy - size 0")
			re, err = s3.Get(op.Filename)
			cached = false
		}

		ext = fname[strings.LastIndex(fname, ".")+1:]
	} else {

		if srcDir != "" {
			// read file from URL
			resp, err := http.Get(srcDir + op.Filename)

			if err != nil {
				log.Println("Failed to open url " + srcDir + op.Filename)
			} else {
				re, err = ioutil.ReadAll(resp.Body)

				if err != nil {
					log.Printf("Failed to read stream " + srcDir)
				}
			}

		} else {
			// read file from bucket
			re, err = s3.Get(op.Filename)
			if err != nil {
				log.Printf("Failed to open bucket " + op.Filename)
			}
		}

	}

	// if we failed to read return message
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Failed to read source object " + srcDir + op.Filename))
	} else {

		if strings.Contains(r.UserAgent(), "Chrome") && op.Format == 0 && params["convert"] == "on" {
			op.Format = 2
		}

		if !cached {
			log.Println("new request " + op.Filename)

			// check image type before
			imtype := image.DetermineImageTypeName(&re)

			if imtype == "" {
				log.Println("image type for request " + op.Filename + " unknown")
				w.Header().Set("Content-Type", "text/plain")
				w.Write([]byte("Source image is not an image "))

			}
			// scale image
			res = image.Resize(re, op)

			// result size
			size = len(res)

			// store only images > 0
			if size > 0 {
				// extension guess
				ext = image.DetermineImageTypeName(&res)

				if ext == "" {

					ext = strings.ToLower(op.Filename[strings.LastIndex(op.Filename, ".")+1:])
				}

				go s3.Put(res, cacheDir+op.Hash+"."+ext, ext)

				etag = helper.GetMD5HashB(res)

				rwrap.HMSet(keyObject, map[string]string{
					"fname": cacheDir + op.Hash + "." + ext,
					"url":   r.URL.Path,
					"md5":   etag,
					"date":  t.Format(time.RFC1123),
				})

				// store size of asset in hash
				rwrap.HINCRBY(keyObject, "size", size)
			} else {
				// can't scale - pass oryginal
				res = re
			}
		}

		/*
			referrer := r.Referer()
			maxl := len(referrer)
			if maxl > 1023 {
				maxl = 1023
			}
			referrer = referrer[:maxl]

			// logging
			go sdb.AddU(vars["user"], map[string]string{
				"origin":  vars["origin"],
				"url":     r.URL.Path,
				"ua":      r.UserAgent(),
				"when":    t.Format(time.RFC3339),
				"size":    strconv.Itoa(len(res)),
				"referer": referrer,
			})*/

		// date header
		w.Header().Set("Date", t.Format(time.RFC1123))

		// lmod
		w.Header().Set("Last-Modified", lmod)

		w.Header().Set("Content-Type", helper.GetImageType(ext))
		w.Header().Set("Content-Length", strconv.Itoa(size))
		w.Header().Set("ETag", "\""+etag+"\"")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.Write(res)
		res = nil

	}
}

func hello(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/plain")
	//	w.Header().Set("Content-Length", strconv.Itoa(len(res)))
	w.Write([]byte("I'm allive"))

	//	w.Write([]byte("hello!"))
}

func handleStatus(w http.ResponseWriter, r *http.Request) {

	if image.LastError != "" {
		w.WriteHeader(204)
		w.Write([]byte(image.LastError))
		fmt.Print(w, image.LastError)
	} else {
		w.WriteHeader(200)
		w.Write([]byte("I'm OK: "))
		w.Write([]byte(version))
	}

}

func purge(surl string) {

	// encode new query
	v := url.Values{}
	v.Set("a", "zone_file_purge")
	v.Set("tkn", gConfig["cloudflare.key"])
	v.Set("email", gConfig["cloudflare.email"])
	v.Set("z", gConfig["cloudflare.domain"])
	v.Set("url", surl)

	resp, err := http.Get("https://www.cloudflare.com/api_json.html?" + v.Encode())

	if err != nil {
		// handle error
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
	/*
	   	curl https://www.cloudflare.com/api_json.html \
	     -d 'a=zone_file_purge' \
	     -d 'tkn=8afbe6dea02407989af4dd4c97bb6e25' \
	     -d 'email=sample@example.com' \
	     -d 'z=example.com' \
	     -d 'url=http://www.example.com/style.css'
	*/
}
