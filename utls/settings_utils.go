package utls

import (
	"strings"
)

//列表参考 https://cloud.d.xiaomi.net/ 数据流服务中topic对应ClusterName
var lcs_talos_map = map[string]string{
	"vru1": "awsvru0-talos",  //AWS-俄罗斯
	"fr1":  "awsde0m-talos",  //AWS-德国1区
	"or1":  "awsusor1-talos", //AWS-美西1区
	"sgp1": "awssgp1-talos",  //AWS-新加坡1区
	"sgp2": "awssgp1-talos",  //AWS-新加坡1区
	"c3":   "cnbj1-talos",    //小米-C3
	"lugu": "cnbj1-talos",    //小米-C3
	"c4":   "cnbj1-talos",    //小米-C3

}

var lcs_enabled = map[string]bool{
	"vru1":    true,
	"fr1":     true,
	"c3":      false,
	"lugu":    false,
	"sgp1":    true,
	"sgp2":    true,
	"or1":     true,
	"staging": false,
	"c4":      false,
}

//LCS服务为防止无法寻找到topic，需要对category添加前缀
func FixCategoryByIdc(idc, category string) string {
	prefix, ok := lcs_talos_map[idc]
	if !ok {
		return category
	} else {
		return strings.Join([]string{prefix, category}, "#")
	}
}

//根据idc判断是否启用LCS服务
func CheckIfUseLCSByIdc(idc string) bool {
	return lcs_enabled[idc]
}
