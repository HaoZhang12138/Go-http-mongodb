package httpfunction

import "github.com/bitly/go-simplejson"

import (
	"net/http"
	"dao"
	"encoding/json"
	"errors"
	"strings"
	"strconv"
	"log"
	"os"
	"io"
)

type Datatablesdata struct {
	Data []interface{}  `json:"data"`
	Files uploadfile_tmp `json:"files,omitempty"`
	Upload uploadid `json:"upload,omitempty"`
}

type uploadid struct {
	Id string `json:"id"`
}

type uploadfile_tmp struct {
	Files *simplejson.Json `json:"files"`
}

const UPLOAD_DIR = "/home/zh/GoPro/firstRedis/html/uploads"
const PreWebPath = "/uploads/"

func loadUserInfoForDatatables(r *http.Request,id []string)(ret []dao.EditDataTables){

	ret = make([]dao.EditDataTables, 0)
	for i := range id {
		var info *dao.UserLoginInfo
		info = new(dao.UserLoginInfo)
		info.User = r.FormValue("data[" + id[i] + "][user]")
		info.Pass = r.FormValue("data[" + id[i] + "][pass]")
		info.Name = r.FormValue("data[" + id[i] + "][name]")
		info.Age  = r.FormValue("data[" + id[i] + "][age]")
		info.Tel  = r.FormValue("data[" + id[i] + "][tel]")
		info.Sex  = r.FormValue("data[" + id[i] + "][sex]")
		info.Id   = r.FormValue("data[" + id[i] + "][_id]")
		info.FileId = r.FormValue("data[" + id[i] + "][fileid]")

		if info.Id == "" {
			info.Id = strconv.Itoa(dao.GetNextId())
		}
		ret = append(ret, info)
	}
	return
}


func GetDataTableId(r *http.Request)(id []string, err error ){

	action := r.FormValue("action")
	id = make([]string, 0)
	if action == "create" {
		id = append(id, "0")
	}else {
		id = make([]string, 0)
		err = errors.New("failed to parse id")
		for k,v := range r.Form{
			if strings.Contains(k,"_id"){
				id = append(id, v[0])
				err = nil
			}
		}
		if err != nil {
			log.Println("failed to get datatables id")
		}
	}
	return
}

func Getpostfile(w http.ResponseWriter, r *http.Request) (uploadid string, err error){
	var filetmp dao.Uploadfile
	file,handler, err := r.FormFile("upload")
	if err != nil {
		log.Println("failed to get formfile")
		return
	}
	defer file.Close()

	filetmp.Filename = handler.Filename
	t, err := os.Create(UPLOAD_DIR + "/" + filetmp.Filename)
	defer t.Close()
	if err != nil {
		log.Println("fail to create file")
		return
	}
	_, err = io.Copy(t, file)
	if err != nil {
		log.Println("failed to copy file")
		return
	}

	filetmp.Systempath = UPLOAD_DIR + "/" + filetmp.Filename
	filetmp.Webpath = PreWebPath + filetmp.Filename

	fileinfo, err := os.Stat(filetmp.Systempath)
	if err != nil {
		log.Println("failed to get the file station")
		return
	}
	filetmp.Filesize = strconv.Itoa(int(fileinfo.Size()))
	filetmp.Id = strconv.Itoa(dao.GetNextId())

	//database
	filetmp.Insert()

	uploadid = filetmp.Id

	return
}


func Createdatatablesline(w http.ResponseWriter,r *http.Request, id []string) (res []dao.EditDataTables, err error) {
	res = loadUserInfoForDatatables(r, id)
	for i := range res {
		err = res[i].Insert()
		if err != nil {
			log.Println("failed to create datatables row")
			return
		}
	}

	return
}

func Editdatatablesline(w http.ResponseWriter,r *http.Request, id []string) (res []dao.EditDataTables,err error) {

	res = loadUserInfoForDatatables(r, id)

	for i := range res {

		var filetmp dao.Uploadfile
		filetmp.Id, err = res[i].GetfileId()
		NowFileId := r.FormValue("data[" + id[i] + "][fileid]")
		if err != nil {
			log.Println("failed to get fileId in func Editdatatablesline")
			return
		}

		if filetmp.Id != "" && filetmp.Id != NowFileId{
			//if need delete local file you should have this
			//err = filetmp.LoadUploadfile()
			//if err != nil {
			//	log.Println("failed to load uploadfile in func deldatatablesline")
			//	return
			// }

			err = filetmp.Remove()
			if err != nil {
				log.Println("failed to remove uploadfile in func deldatatablesline")
				return
			}
		}

		err = res[i].Update()
		if err != nil {
			log.Println("failed to update datatables line")
			return
		}

	}
	return
}


func Deldatatablesline(w http.ResponseWriter,r *http.Request, id []string) (err error) {

	res := loadUserInfoForDatatables(r, id)

	for i := range res {
		var filetmp dao.Uploadfile
		filetmp.Id, err = res[i].GetfileId()
		if err != nil {
			log.Println("failed to get fileId in func Deldatatablesline")
			return
		}
		if filetmp.Id != ""{
			//if need delete local file you should have this
			//err = filetmp.LoadUploadfile()
			//if err != nil {
			//	log.Println("failed to load uploadfile in func deldatatablesline")
			//	return
			//}

			err = filetmp.Remove()
			if err != nil {
				log.Println("failed to remove uploadfile in func deldatatablesline")
				return
			}
		}

		err = res[i].Remove()
		if err != nil {
			log.Println("failed to delete datatables line")
			return
		}
	}

	return
}


func ViewHandle(w http.ResponseWriter, r *http.Request){

	var tmp Datatablesdata
	var err error
	if r.Method == "POST"{
		if r.ContentLength != 0 {
			action := r.FormValue("action")
			if action == "upload" {
				tmp.Upload.Id, err = Getpostfile(w, r)
				if err != nil {
					log.Println(err.Error(), " failed to action upload in func ViewHandle")
					return
				}

			} else {
				id, _ := GetDataTableId(r)
				if action == "create"  || action == "edit"{
					var res []dao.EditDataTables
					if action == "create" {
						res, _ = Createdatatablesline(w,r,id)
					}else {
						res, _ = Editdatatablesline(w,r,id)
					}

					for i := range res {
						tmp.Data = append(tmp.Data, res[i].GetOneById())
					}


				}else if action == "remove"{
					Deldatatablesline(w,r,id)
				}
			}
		}
	} else if r.Method == "GET" {
		tmp.Data, err = new(dao.UserLoginInfo).GetAll()
		if err != nil {
			log.Println(err.Error(), " failed to GET in func ViewHandle")
			return
		}
	}

	tmp.Files.Files = simplejson.New()
	res, err := dao.GetAllUploadfile()
	if err != nil {
		log.Println(err.Error(), " failed to GetAllUploadfile in func ViewHandle")
		return
	}
	len := len(res)
	for i := 0; i < len; i++ {
		tmp.Files.Files.Set(res[i].Id, res[i])
	}

	t, err := json.Marshal(tmp)
	if err != nil {
		log.Println(err.Error(), " failed to marshal to json in func ViewHandle")
		return
	}
	w.Write(t)
}


