package handle

import (
	"fmt"
	"strings"
	"unicode"

	core "github.com/nguyendev/zago/internal/core"
	"github.com/nguyendev/zago/internal/util"
	"github.com/nguyendev/zago/internal/worker"
)

var bankMap = map[string]map[string]string{
	"422589": {"bank": "CIMB", "name": "Ngân hàng TNHH MTV CIMB Việt Nam (CIMB)"},
	"458761": {"bank": "HSBC", "name": "Ngân hàng TNHH MTV HSBC (Việt Nam) (HSBC)"},
	"546034": {"bank": "CAKE", "name": "Ngân hàng số CAKE by VPBank (CAKE)"},
	"546035": {"bank": "Ubank", "name": "Ngân hàng số Ubank by VPBank (Ubank)"},
	"668888": {"bank": "KBank", "name": "Ngân hàng Đại chúng TNHH Kasikornbank (KBank)"},
	"796500": {"bank": "DBSBank", "name": "DBS Bank Ltd - Chi nhánh TP. Hồ Chí Minh (DBSBank)"},
	"801011": {"bank": "Nonghyup", "name": "Ngân hàng Nonghyup - Chi nhánh Hà Nội (Nonghyup)"},
	"970400": {"bank": "SaigonBank", "name": "Ngân hàng TMCP Sài Gòn Công Thương (SaigonBank)"},
	"970403": {"bank": "Sacombank", "name": "Ngân hàng TMCP Sài Gòn Thương Tín (Sacombank)"},
	"970405": {"bank": "Agribank", "name": "Ngân hàng Nông nghiệp và Phát triển Nông thôn Việt Nam (Agribank)"},
	"970406": {"bank": "DongABank", "name": "Ngân hàng TMCP Đông Á (DongABank)"},
	"970407": {"bank": "Techcombank", "name": "Ngân hàng TMCP Kỹ thương Việt Nam (Techcombank)"},
	"970408": {"bank": "GPBank", "name": "Ngân hàng Thương mại TNHH MTV Dầu Khí Toàn Cầu (GPBank)"},
	"970409": {"bank": "BacABank", "name": "Ngân hàng TMCP Bắc Á (BacABank)"},
	"970410": {"bank": "StandardChartered", "name": "Ngân hàng Standard Chartered Việt Nam (Standard Chartered)"},
	"970412": {"bank": "PVcomBank", "name": "Ngân hàng TMCP Đại Chúng Việt Nam (PVcomBank)"},
	"970414": {"bank": "Oceanbank", "name": "Ngân hàng Thương mại TNHH MTV Đại Dương (Oceanbank)"},
	"970415": {"bank": "VietinBank", "name": "Ngân hàng TMCP Công thương Việt Nam (VietinBank)"},
	"970416": {"bank": "ACB", "name": "Ngân hàng TMCP Á Châu (ACB)"},
	"970418": {"bank": "BIDV", "name": "Ngân hàng TMCP Đầu tư và Phát triển Việt Nam (BIDV)"},
	"970419": {"bank": "NCB", "name": "Ngân hàng TMCP Quốc Dân (NCB)"},
	"970421": {"bank": "VRB", "name": "Ngân hàng Liên doanh Việt - Nga (VRB)"},
	"970422": {"bank": "MBBank", "name": "Ngân hàng TMCP Quân đội (MBBank)"},
	"970423": {"bank": "TPBank", "name": "Ngân hàng TMCP Tiên Phong (TPBank)"},
	"970424": {"bank": "ShinhanBank", "name": "Ngân hàng TNHH MTV Shinhan Việt Nam (ShinhanBank)"},
	"970425": {"bank": "ABBANK", "name": "Ngân hàng TMCP An Bình (ABBANK)"},
	"970426": {"bank": "MSB", "name": "Ngân hàng TMCP Hàng Hải (MSB)"},
	"970427": {"bank": "VietABank", "name": "Ngân hàng TMCP Việt Á (VietABank)"},
	"970428": {"bank": "NamABank", "name": "Ngân hàng TMCP Nam Á (NamABank)"},
	"970429": {"bank": "SCB", "name": "Ngân hàng TMCP Sài Gòn (SCB)"},
	"970430": {"bank": "PGBank", "name": "Ngân hàng TMCP Xăng dầu Petrolimex (PGBank)"},
	"970431": {"bank": "Eximbank", "name": "Ngân hàng TMCP Xuất Nhập khẩu Việt Nam (Eximbank)"},
	"970432": {"bank": "VPBank", "name": "Ngân hàng TMCP Việt Nam Thịnh Vượng (VPBank)"},
	"970433": {"bank": "VietBank", "name": "Ngân hàng TMCP Việt Nam Thương Tín (VietBank)"},
	"970434": {"bank": "IndovinaBank", "name": "Ngân hàng TNHH Indovina (IndovinaBank)"},
	"970436": {"bank": "Vietcombank", "name": "Ngân hàng TMCP Ngoại Thương Việt Nam (Vietcombank)"},
	"970437": {"bank": "HDBank", "name": "Ngân hàng TMCP PT Thành phố Hồ Chí Minh (HDBank)"},
	"970438": {"bank": "BaoVietBank", "name": "Ngân hàng TMCP Bảo Việt (BaoVietBank)"},
	"970439": {"bank": "PublicBank", "name": "Ngân hàng TNHH MTV Public Việt Nam (PublicBank)"},
	"970440": {"bank": "SeABank", "name": "Ngân hàng TMCP Đông Nam Á (SeABank)"},
	"970441": {"bank": "VIB", "name": "Ngân hàng TMCP Quốc tế Việt Nam (VIB)"},
	"970442": {"bank": "HongLeong", "name": "Ngân hàng TNHH MTV Hong Leong Việt Nam (HongLeong)"},
	"970443": {"bank": "SHB", "name": "Ngân hàng TMCP Sài Gòn - Hà Nội (SHB)"},
	"970444": {"bank": "CBBank", "name": "Ngân hàng TM TNHH MTV Xây dựng Việt Nam (CBBank)"},
	"970446": {"bank": "COOPBANK", "name": "Ngân hàng Hợp tác xã Việt Nam (COOPBANK)"},
	"970448": {"bank": "OCB", "name": "Ngân hàng TMCP Phương Đông (OCB)"},
	"970449": {"bank": "LienVietPostBank", "name": "Ngân hàng TMCP Bưu Điện Liên Việt (LienVietPostBank)"},
	"970452": {"bank": "KienLongBank", "name": "Ngân hàng TMCP Kiên Long (KienLongBank)"},
	"970454": {"bank": "VietCapitalBank", "name": "Ngân hàng TMCP Bản Việt (VietCapitalBank)"},
	"970455": {"bank": "IBKHN", "name": "Ngân hàng Công nghiệp Hàn Quốc - Chi nhánh Hà Nội (IBKHN)"},
	"970456": {"bank": "IBKHCM", "name": "Ngân hàng Công nghiệp Hàn Quốc - Chi nhánh TP. Hồ Chí Minh (IBKHCM)"},
	"970457": {"bank": "Woori", "name": "Ngân hàng TNHH MTV Woori Việt Nam (Woori)"},
	"970458": {"bank": "UnitedOverseas", "name": "Ngân hàng United Overseas - Chi nhánh TP. Hồ Chí Minh (United Overseas)"},
	"970462": {"bank": "KookminHN", "name": "Ngân hàng Kookmin - Chi nhánh Hà Nội (KookminHN)"},
	"970463": {"bank": "KookminHCM", "name": "Ngân hàng Kookmin - Chi nhánh Tp. Hồ Chí Minh (KookminHCM)"},
}

func normBankKey(s string) string {
	var b strings.Builder
	for _, ch := range strings.TrimSpace(strings.ToLower(s)) {
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

func resolveBinCode(bank string) (string, error) {
	bank = strings.TrimSpace(bank)
	if bank == "" {
		return "", fmt.Errorf("missing bank")
	}
	if _, ok := bankMap[bank]; ok {
		return bank, nil
	}
	needle := normBankKey(bank)
	for bin, info := range bankMap {
		if normBankKey(info["bank"]) == needle || normBankKey(info["name"]) == needle {
			return bin, nil
		}
	}
	return "", fmt.Errorf("unknown bank: %s", bank)
}

func (s *SendAPI) SendCardBank(msgData *worker.MessageObject, bankNum string, nameAccBank string, bank string, threadID string, threadType core.ThreadType) (any, error) {
	if strings.TrimSpace(bankNum) == "" {
		return nil, fmt.Errorf("missing bank account number")
	}
	if strings.TrimSpace(nameAccBank) == "" {
		return nil, fmt.Errorf("missing account owner name")
	}
	binCode, err := resolveBinCode(bank)
	if err != nil {
		return nil, err
	}
	destType := 0
	if threadType == core.GROUP {
		destType = 1
	}
	form, err := s.EncodedForm(map[string]any{"binBank": binCode, "numAccBank": bankNum, "nameAccBank": nameAccBank, "cliMsgId": util.AsString(msgData.Get("cliMsgId")), "tsMsg": map[bool]any{true: threadID, false: msgData.Get("ts")}[destType == 1], "destUid": threadID, "destType": destType})
	if err != nil {
		return nil, err
	}
	data, err := s.PostJSON("https://zimsg.chat.zalo.me/api/transfer/card", s.Query(map[string]any{"zpw_ver": 648, "zpw_type": s.APILoginType}), form)
	if err != nil {
		return nil, err
	}
	return s.ParseRaw(data)
}
