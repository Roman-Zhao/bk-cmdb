/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package logics

import (
	"fmt"
	"strings"

	"configcenter/src/common"
	"configcenter/src/common/blog"
	"configcenter/src/common/language"
	"configcenter/src/common/mapstr"
	"configcenter/src/common/metadata"
	"configcenter/src/common/util"

	"github.com/rentiansheng/xlsx"
)

const (
	fieldTypeBoolTrue  = "true"
	fieldTypeBoolFalse = "false"
)

// getFieldsIDIndexMap get field property index
func getFieldsIDIndexMap(fields map[string]Property) map[string]int {
	index := 0
	IDNameMap := make(map[string]int)
	for id := range fields {
		IDNameMap[id] = index
		index++
	}
	return IDNameMap
}

// getAssociateName  get getAssociate object name
func getAssociatePrimaryKey(a []interface{}, primaryField []Property) []string {
	vals := []string{}
	for _, valRow := range a {
		mapVal, ok := valRow.(map[string]interface{})
		if ok {
			instMap, ok := mapVal["inst_info"].(map[string]interface{})
			if true == ok {
				var itemVals []string
				for _, field := range primaryField {
					val, _ := instMap[field.ID]
					if nil == val {
						val = ""
					}
					itemVals = append(itemVals, fmt.Sprintf("%v", val))
				}
				vals = append(vals, strings.Join(itemVals, common.ExcelAsstPrimaryKeySplitChar))
			}
		}
	}

	return vals
}

// getEnumNameByID get enum name from option
func getEnumNameByID(id string, items []interface{}) string {
	var name string
	for _, valRow := range items {
		mapVal, ok := valRow.(map[string]interface{})
		if ok {
			enumID, ok := mapVal["id"].(string)
			if true == ok {
				if enumID == id {
					name = mapVal["name"].(string)
				}
			}
		}
	}

	return name
}

// getEnumIDByName get enum name from option
func getEnumIDByName(name string, items []interface{}) string {
	id := name
	for _, valRow := range items {
		mapVal, ok := valRow.(map[string]interface{})
		if ok {
			enumName, ok := mapVal["name"].(string)
			if true == ok {
				if enumName == name {
					id = mapVal["id"].(string)
				}
			}
		}
	}

	return id
}

// getEnumNames get enum name from option
func getEnumNames(items []interface{}) []string {
	var names []string
	for _, valRow := range items {
		mapVal, ok := valRow.(map[string]interface{})
		if ok {

			name, ok := mapVal["name"].(string)
			if ok {
				names = append(names, name)
			}

		}
	}

	return names
}

// getHeaderCellGeneralStyle get excel header general style by C6EFCE,000000
func getHeaderCellGeneralStyle() *xlsx.Style {
	return getCellStyle(common.ExcelHeaderOtherRowColor, common.ExcelHeaderOtherRowFontColor)
}

// getHeaderFirstRowCellStyle
func getHeaderFirstRowCellStyle(isRequire bool) *xlsx.Style {
	if isRequire {
		return getCellStyle(common.ExcelHeaderFirstRowColor, common.ExcelHeaderFirstRowRequireFontColor)
	}

	return getCellStyle(common.ExcelHeaderFirstRowColor, common.ExcelHeaderFirstRowFontColor)
}

// getCellStyle get cell style from fgColor and fontcolor
func getCellStyle(fgColor, fontColor string) *xlsx.Style {
	style := xlsx.NewStyle()
	style.Fill = *xlsx.DefaultFill()
	style.Font = *xlsx.DefaultFont()
	style.ApplyFill = true
	style.ApplyFont = true
	style.ApplyBorder = true

	style.Border = *xlsx.NewBorder("thin", "thin", "thin", "thin")
	style.Border.BottomColor = common.ExcelCellDefaultBorderColor
	style.Border.TopColor = common.ExcelCellDefaultBorderColor
	style.Border.LeftColor = common.ExcelCellDefaultBorderColor
	style.Border.RightColor = common.ExcelCellDefaultBorderColor

	style.Fill.FgColor = fgColor
	style.Fill.PatternType = "solid"

	style.Font.Color = fontColor

	return style
}

// addExtFields  add extra fields,
func addExtFields(fields map[string]Property, extFields map[string]string, extFieldKey []string) map[string]Property {
	excelColIndex := 0
	for _, extFieldID := range extFieldKey {
		fields[extFieldID] = Property{
			ID:            "",
			Name:          extFields[extFieldID],
			NotObjPropery: true,
			ExcelColIndex: excelColIndex,
		}
		excelColIndex++
	}

	for _, field := range fields {
		if excelColIndex < field.ExcelColIndex {
			excelColIndex = field.ExcelColIndex
		}
	}

	return fields
}

func replaceEnName(rid string, rowMap mapstr.MapStr, usernameMap map[string]string, propertyList []string,
	defLang language.DefaultCCLanguageIf) (mapstr.MapStr, error) {
	// propertyList是用户自定义的objuser型的attr名列表
	for _, property := range propertyList {
		if rowMap[property] == nil {
			continue
		}

		userListString, ok := rowMap[property].(string)
		if !ok {
			blog.Errorf("convert variable rowMap[%s] type to string field , rowMap: %v, rowMap type: %T, rid: %s", property, rowMap[property], rowMap[property], rid)
			return nil, fmt.Errorf("convert variable rowMap[%s] type to string field", property)
		}
		userListString = strings.TrimSpace(userListString)
		if userListString == "" {
			continue
		}

		newUserList := []string{}
		enNameList := strings.Split(userListString, ",")
		for _, enName := range enNameList {
			username := usernameMap[enName]
			if username == "" {
				// return the original user name and remind that the user is nonexistent in '()'
				username = fmt.Sprintf("%s(%s)", enName, defLang.Language("nonexistent_user"))
			}
			newUserList = append(newUserList, username)
		}
		rowMap[property] = strings.Join(newUserList, ",")
	}

	return rowMap, nil
}

// setExcelCellIgnore set the excel cell to be ignored
func setExcelCellIgnored(sheet *xlsx.Sheet, style *xlsx.Style, row int, col int) {
	cell := sheet.Cell(row, col)
	cell.Value = common.ExcelCellIgnoreValue
	cell.SetStyle(style)
}

// replaceDepartmentFullName replace attribute organization's id by fullname in export excel
func replaceDepartmentFullName(rid string, rowMap mapstr.MapStr, org []metadata.DepartmentItem, propertyList []string,
	defLang language.DefaultCCLanguageIf) (mapstr.MapStr, error) {

	orgMap := make(map[int64]string)
	for _, item := range org {
		orgMap[item.ID] = item.FullName
	}

	for _, property := range propertyList {
		orgIDInterface, exist := rowMap[property]
		if !exist || orgIDInterface == nil {
			continue
		}

		orgIDList, ok := orgIDInterface.([]interface{})
		if !ok {
			blog.Errorf("rowMap[%s] type to array failed, rowMap: %v, rowMap type: %T, rid: %s", property,
				rowMap[property], rowMap[property], rid)
			return nil, fmt.Errorf("convert variable rowMap[%s] type to int array failed", property)
		}

		orgName := make([]string, 0)
		for _, orgID := range orgIDList {
			id, err := util.GetInt64ByInterface(orgID)
			if err != nil {
				blog.Errorf("convert orgID[%v] to int64 failed, type: %T, err: %v, rid: %s", orgID, orgID, err, rid)
				return nil, fmt.Errorf("convert variable orgID[%v] type to int64 failed", orgID)
			}

			name, exist := orgMap[id]
			if !exist {
				blog.Errorf("orgnization[%d] does no exist, rid: %s", id, rid)
				orgName = append(orgName, fmt.Sprintf("[%d]%s", id, defLang.Language("nonexistent_org")))
				continue
			}

			orgName = append(orgName, fmt.Sprintf("[%d]%s", id, name))
		}
		rowMap[property] = strings.Join(orgName, ",")
	}

	return rowMap, nil
}
