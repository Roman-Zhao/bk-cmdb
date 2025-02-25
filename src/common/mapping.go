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

package common

// GetInstNameField returns the inst name field
func GetInstNameField(objID string) string {
	switch objID {
	case BKInnerObjIDBizSet:
		return BKBizSetNameField
	case BKInnerObjIDApp:
		return BKAppNameField
	case BKInnerObjIDSet:
		return BKSetNameField
	case BKInnerObjIDModule:
		return BKModuleNameField
	case BKInnerObjIDObject:
		return BKInstNameField
	case BKInnerObjIDHost:
		return BKHostNameField
	case BKInnerObjIDProc:
		return BKProcNameField
	case BKInnerObjIDPlat:
		return BKCloudNameField
	case BKTableNameInstAsst:
		return BKFieldID
	default:
		if IsObjectInstAsstShardingTable(objID) {
			return BKFieldID
		}
		return BKInstNameField
	}
}

// GetInstIDField get primary key of object's collection/table
func GetInstIDField(objType string) string {
	switch objType {
	case BKInnerObjIDBizSet:
		return BKBizSetIDField
	case BKInnerObjIDApp:
		return BKAppIDField
	case BKInnerObjIDSet:
		return BKSetIDField
	case BKInnerObjIDModule:
		return BKModuleIDField
	case BKInnerObjIDObject:
		return BKInstIDField
	case BKInnerObjIDHost:
		return BKHostIDField
	case BKInnerObjIDProc:
		return BKProcIDField
	case BKInnerObjIDPlat:
		return BKCloudIDField
	case BKTableNameInstAsst:
		return BKFieldID
	case BKTableNameServiceInstance:
		return BKFieldID
	case BKTableNameServiceTemplate:
		return BKFieldID
	case BKTableNameProcessTemplate:
		return BKFieldID
	case BKTableNameProcessInstanceRelation:
		return BKProcessIDField
	default:
		if IsObjectInstAsstShardingTable(objType) {
			return BKFieldID
		}
		return BKInstIDField
	}
}

func GetObjByType(objType string) string {
	switch objType {
	case BKInnerObjIDBizSet, BKInnerObjIDApp, BKInnerObjIDSet,
		BKInnerObjIDModule, BKInnerObjIDProc,
		BKInnerObjIDHost, BKInnerObjIDPlat:
		return objType
	default:
		return BKInnerObjIDObject
	}
}

func IsInnerModel(objType string) bool {
	return GetObjByType(objType) != BKInnerObjIDObject
}

// IsInnerMainlineModel judge if the object type is inner mainline model
func IsInnerMainlineModel(objType string) bool {
	switch objType {
	case BKInnerObjIDApp, BKInnerObjIDSet, BKInnerObjIDModule:
		return true
	default:
		return false
	}
}
