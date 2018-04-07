package ramdominfo

var (
	// PhoneModelDataBase 手机型号库
	PhoneModelDataBase = []string{
		"HUAWEI MT7-AL00", //华为mate7
		"HUAWEI NXT-AL10", //华为mate8
		"HUAWEI CAZ-AL10", //华为nova
		"BLN-AL10",        //荣耀6X
		"BND-AL10",        //荣耀7X
		"KNT-AL10",        //荣耀V8
		"NWI-AL10",        //nova2s
		"TA-1000",         //Nokia 6
		"TA-1041",         //Nokia 7
		"OPPO R11",
		"VIVO X20",
		"VIVO X20A",
		"SM-G9650",      //Samsung Galasy S9+
		"SM-G9600",      //Samsung Galasy S9
		"SM-G960F",      //Galaxy S9 Dual SIM
		"SM-G965F",      //Galaxy S9+ Dual SIM
		"SM-G9500",      //Samsung Galasy S8
		"SM-G9250",      //Samsung Galasy S7 Edge
		"SM-G7200",      //Samsung Galasy Grand Max
		"SM-N9500",      //Samsung Galasy Note8
		"SM-N9108V",     //Samsung Galasy Note4
		"HTC U-1w",      //HTC U Ultra
		"Lenovo K52e78", //Lenovo K5 Note
		"ZUK Z2121",     //ZUK Z2 Pro
		"MiTV2S-48",     //小米电视2s
		"Redmi 3S",      //红米3s
		"Mi A1",         //MI androidone
		"Mi 6",
		"Nexus 4",
		"G8142",         //索尼Xperia XZ Premium G8142
		"G8342",         //索尼Xperia XZ1
		"NX563J",        //努比亚Z17
		"S3",            //佳域S3
		"STV100-1",      //黑莓Priv
		"ONEPLUS A5010", //一加5T
		"GRA-A0",        //Coolpad Cool Play 6C
	}
)

// SumIMEI 根据key计算出imei
func SumIMEI(key string) uint64 {
	var hash uint64 = 53202347234687234
	for k := range key {
		hash += (hash << 5) + uint64(key[k])
	}
	hash %= uint64(1e15)
	if hash < 1e14 {
		hash += 1e14
	}
	return hash
}

// GetPhoneModel 根据key, 从PhoneModelDataBase中取出手机型号
func GetPhoneModel(key string) string {
	var hash uint64 = 2134
	for k := range key {
		hash += (hash << 4) + uint64(key[k])
	}
	hash %= uint64(len(PhoneModelDataBase))
	return PhoneModelDataBase[int(hash)]
}
