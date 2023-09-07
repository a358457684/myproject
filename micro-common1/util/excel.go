package util

import (
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"time"
)

const DefaultSheet = "Sheet1"

func ExportExcel(file *excelize.File, rowFrom int, mainHeader string, headers, attrs []string, pdsList []map[string]interface{}) *excelize.File {
	if headers == nil || len(headers) != len(attrs) {
		return file
	}

	if file == nil {
		file = excelize.NewFile()
		file.NewSheet(DefaultSheet)
	} else if file.GetSheetIndex(DefaultSheet) == 0 {
		file.NewSheet(DefaultSheet)
	}

	left := getcell(rowFrom, 1)
	right := getcell(rowFrom, len(headers))
	if left == "" || right == "" {
		return file
	}
	file.MergeCell(DefaultSheet, left, right)
	file.SetCellValue(DefaultSheet, left, mainHeader)
	style, _ := file.NewStyle(`{
	"alignment":
    {
        "horizontal": "center"
	},
    "font":
    {
        "bold": true,
        "family": "黑体",
        "size": 20
    }
	}`)
	file.SetCellStyle(DefaultSheet, left, right, style)

	titlestyle, _ := file.NewStyle(`{
"fill":{"type":"pattern","color":["#00CCFF"],"pattern":1},
 "border": [
    {
        "type": "left",
		"color":"#000000",
        "style": 2
    },
    {
        "type": "top",
		"color":"#000000",
        "style": 2
    },
    {
        "type": "bottom",
		"color":"#000000",
        "style": 2
    },
    {
        "type": "right",
		"color":"#000000",
        "style": 2
    }
],
	"alignment":
    {
        "horizontal": "center"
	},
    "font":
    {
        "bold": true,
        "family": "黑体",
        "size": 14
    }
	}`)

	for key, item := range headers {
		file.SetCellValue(DefaultSheet, getcell(rowFrom+1, key+1), item)
		if rowFrom == 1 {
			file.SetColWidth(DefaultSheet, getcol(key+1), getcol(key+1), float64(len(item)*2))
		}
		file.SetCellStyle(DefaultSheet, getcell(rowFrom+1, key+1), getcell(rowFrom+1, key+1), titlestyle)
	}

	cellstyle, _ := file.NewStyle(`{

 "border": [
    {
        "type": "left",
		"color":"#000000",
        "style": 1
    },
    {
        "type": "top",
		"color":"#000000",
        "style": 1
    },
    {
        "type": "bottom",
		"color":"#000000",
        "style": 1
    },
    {
        "type": "right",
		"color":"#000000",
        "style": 1
    }
],
	"alignment":
    {
        "horizontal": "center"
	}
	}`)

	count := 0
	for _, pd := range pdsList {
		count++

		for key, item := range attrs {
			file.SetCellStyle(DefaultSheet, getcell(rowFrom+1+count, key+1), getcell(rowFrom+1+count, key+1), cellstyle)
			if item == "appear_date" || item == "finish_date" {
				if pd[item] == nil {
					continue
				}
				if value, isok := pd[item].(time.Time); isok {
					if value.Unix() == -62135625943 {
						continue
					}
					pd[item] = value.Format("2006-01-02 15:04:05")
				}
			}

			file.SetCellValue(DefaultSheet, getcell(rowFrom+1+count, key+1), pd[item])
		}
	}
	return file
}

// 通用的导出模版，适用于数据比较简单，只使用到A~Z的列，导出超过26列会有问题
func GeneralExcelExport(title string, headers []string, dataLists [][]interface{}, colWidths map[string]float64) (*excelize.File, error) {

	excel := excelize.NewFile()
	excel.NewSheet(title)
	excel.DeleteSheet(DefaultSheet)

	// 样式
	lastColumn := fmt.Sprintf("%c", 64+len(headers))
	_ = excel.SetColWidth(title, "A", lastColumn, 22)
	for colName, width := range colWidths {
		_ = excel.SetColWidth(title, colName, colName, width)
	}
	alignment := excelize.Alignment{Horizontal: "center", Vertical: "center"}
	titleFont := excelize.Font{Bold: true, Family: "微软雅黑", Size: 20}
	headerFont := excelize.Font{Family: "微软雅黑", Size: 12, Color: "#FFFFFF"}
	headerFill := excelize.Fill{Type: "pattern", Color: []string{"#808080"}, Pattern: 1}
	dataFont := excelize.Font{Family: "微软雅黑", Size: 10}

	// 标题头
	titleStyle, _ := excel.NewStyle(&excelize.Style{Font: &titleFont, Alignment: &alignment})
	_ = excel.SetRowHeight(title, 1, 30)
	_ = excel.SetCellStyle(title, "A1", "A1", titleStyle)
	_ = excel.MergeCell(title, "A1", lastColumn+"1")
	_ = excel.SetCellValue(title, "A1", title)

	// 标题
	headerStyle, _ := excel.NewStyle(&excelize.Style{Font: &headerFont, Fill: headerFill, Alignment: &alignment})
	_ = excel.SetCellStyle(title, "A2", lastColumn+"2", headerStyle)
	_ = excel.SetRowHeight(title, 2, 18)
	for index, value := range headers {
		cell := fmt.Sprintf("%c2", 65+index)
		_ = excel.SetCellValue(title, cell, value)
	}

	// 数据
	dataStyle, _ := excel.NewStyle(&excelize.Style{Font: &dataFont, Alignment: &alignment})
	_ = excel.SetCellStyle(title, "A3", fmt.Sprintf("%s%d", lastColumn, 2+len(dataLists)), dataStyle)
	for index, dataList := range dataLists {
		for valueIndex, value := range dataList {
			cell := fmt.Sprintf("%c%d", 65+valueIndex, 3+index)
			err := excel.SetCellValue(title, cell, value)
			if err != nil {
				return excel, err
			}
		}
	}

	return excel, nil
}

func getcell(row int, column int) string {
	if column > 26 || column <= 0 || row <= 0 {
		return ""
	}
	return fmt.Sprintf("%c%d", column+64, row)
}

func getcol(column int) string {
	if column > 26 || column <= 0 {
		return ""
	}
	return fmt.Sprintf("%c", column+64)
}
