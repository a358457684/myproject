package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

const (
	baseColumn = "id,create_by,create_date,update_by,update_date,del_flag"
	destDir    = "../../common/biz/"
)

func GetNewSqlFilePath(dir string) string {
	files, _ := ioutil.ReadDir(dir)
	var newFile os.FileInfo
	for _, file := range files {
		if strings.Contains(file.Name(), ".sql") {
			if file.Mode().IsRegular() {
				if newFile == nil || file.ModTime().After(newFile.ModTime()) {
					newFile = file
				}
			}
		}
	}
	return dir + newFile.Name()
}

func SqlToStruct(filePath, newName string, isBase bool) error {
	data, _ := ioutil.ReadFile(filePath)
	tableName := regexp.MustCompile("CREATE TABLE `(.*)`").FindSubmatch(data)[1]
	comment := regexp.MustCompile("COMMENT='(.*)';").FindSubmatch(data)[1]
	var fields string
	if isBase {
		fields += "\tBaseEntity\n"
	}
	columns := regexp.MustCompile("[ ][ ]`(.*)` (\\w*).* COMMENT '(([^\\x00-\\xff]|\\w|\\(|\\)|，|:)*)',").FindAllSubmatch(data, 50)
	hasTime := false
	for _, column := range columns {
		columnName := string(column[1])
		if isBase && strings.Contains(baseColumn, columnName) {
			continue
		}
		var field string
		for _, word := range strings.Split(columnName, "_") {
			field = field + strings.Title(word)
		}
		fieldType := string(column[2])
		switch fieldType {
		case "varchar", "char", "text", "longtext":
			fieldType = "string"
		case "datetime":
			hasTime = true
			fieldType = "time.Time"
		case "float", "double":
			fieldType = "float64"
		case "tinyint", "int", "decimal":
			fieldType = "int"
		case "long":
			fieldType = "int64"
		}
		fields += fmt.Sprintf("\t%s\t%s\t//%s\n", field, fieldType, column[3])
	}
	var newNameTitle string
	for _, world := range strings.Split(newName, "_") {
		newNameTitle += strings.Title(world)
	}

	template := "package entity\n\n"
	if hasTime {
		template += "import \"time\"\n\n"
	}
	template += "//comment\ntype className struct {\nfields}\n\nfunc (className) TableName() string {\n\treturn \"tableName\"\n}\n"
	template = strings.Replace(template, "comment", string(comment), 1)
	template = strings.Replace(template, "className", newNameTitle+"Entity", 2)
	template = strings.Replace(template, "tableName", string(tableName), 1)
	template = strings.Replace(template, "fields", fields, 1)
	fmt.Println(template)
	err := ioutil.WriteFile(destDir+"entity/"+newName+"_entity.go", []byte(template), 0666)

	template = "package dao\n\n"
	err = ioutil.WriteFile(destDir+"dao/"+newName+"_dao.go", []byte(template), 0666)

	return err
}
