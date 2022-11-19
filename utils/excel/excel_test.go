package excel

import (
	"fmt"
	"strconv"
	"testing"

	db "github.com/skirrund/gcloud/datasource"
	"github.com/skirrund/gcloud/utils"
	"github.com/xuri/excelize/v2"
)

func TestXxx(t *testing.T) {
	f, err := excelize.OpenFile("/Users/jerry.shi/Desktop/宸汐健康/益源报价.xlsx")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	sn := f.GetSheetName(0)
	fmt.Println(sn)
	rows, err := f.GetRows(sn)
	if err != nil {
		fmt.Println(err)
		return
	}
	db.InitDataSource(db.Option{
		DSN: "test_admin:CX2021@admin@tcp(rm-uf68i2lk1t03d134r.mysql.rds.aliyuncs.com:3306)/pbm_maindata?charset=utf8mb4&parseTime=True&loc=Local",
	})
	f1, err := excelize.OpenFile("/Users/jerry.shi/Downloads/drug-batch-import.xlsx")
	if err != nil {
		fmt.Println(err)
		return
	}

	defer f1.Close()
	sn1 := f1.GetSheetName(0)
	for i, row := range rows {
		if i == 0 {
			continue
		}
		gdId := row[0]
		commonName := row[1]
		//gName := row[2]
		spec := row[3]
		unit := row[4]
		fac := row[5]
		price := row[7]
		if len(price) > 0 {
			sp, err := utils.NewFromString(price)
			if err != nil {
				fmt.Println(err)
			}
			price = sp.String()
		}
		settlePrice := row[8]
		if len(settlePrice) > 0 {
			sp, err := utils.NewFromString(settlePrice)
			if err != nil {
				fmt.Println(err)
			}
			settlePrice = sp.StringFixed(2)
		}
		lis := row[9]
		cat := row[10] + "," + row[11]
		// err = f1.InsertRow(sn1, i+1)
		// if err != nil {
		// 	fmt.Println(err)
		// 	continue
		// }
		rowNum := int64(i + 1)
		index := strconv.FormatInt(rowNum, 10)
		f1.SetCellValue(sn1, "A"+index, "普药卡")
		f1.SetCellValue(sn1, "B"+index, "宸汐虚拟国药药房1")
		f1.SetCellValue(sn1, "F"+index, commonName)
		f1.SetCellValue(sn1, "D"+index, commonName)
		f1.SetCellValue(sn1, "G"+index, spec)
		f1.SetCellValue(sn1, "H"+index, unit)
		f1.SetCellValue(sn1, "I"+index, fac)
		f1.SetCellValue(sn1, "L"+index, price)
		f1.SetCellValue(sn1, "K"+index, cat)
		f1.SetCellValue(sn1, "M"+index, lis)
		f1.SetCellValue(sn1, "X"+index, gdId)
		f1.SetCellValue(sn1, "N"+index, "9-其他")
		f1.SetCellValue(sn1, "O"+index, "8-其他")
		f1.SetCellValue(sn1, "P"+index, "24-其他")
		if row[10] == "处方药" {
			f1.SetCellValue(sn1, "R"+index, "Y")
		}
		f1.SetCellValue(sn1, "Y"+index, settlePrice)
	}
	err = f1.SaveAs("/Users/jerry.shi/Downloads/drug-batch-import-gd2.xlsx")
	if err != nil {
		fmt.Println(err)
	}
}
