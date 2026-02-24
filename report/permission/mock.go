package permission

import "fmt"

func MockReport(programN, companyN, empTypeN int) Report {
	r := Report{
		Title: "รายงานกำหนดสิทธิผู้ใช้ระบบ",
		Group: "Admin",
		User: User{
			Username: "007",
		},
	}

	for i := 1; i <= programN; i++ {
		r.User.Programs = append(r.User.Programs, CheckItem{
			Label:   fmt.Sprintf("Program %03d - ရှည်လျားသော စာသားကို စမ်းသပ်ပြီး စာမျက်နှာကွဲမှုများကို စောင့်ကြည့်ပါ။", i),
			Checked: true,
		})
	}

	for i := 1; i <= companyN; i++ {
		r.User.Companies = append(r.User.Companies, RowCheck{
			Code:    fmt.Sprintf("%d", i),
			Name:    fmt.Sprintf("ကုမ္ပဏီ %03d (ชื่อยาวเพื่อ wrap) - TigerSoft Group", i),
			Checked: true,
		})
	}

	for i := 1; i <= empTypeN; i++ {
		r.User.EmpTypes = append(r.User.EmpTypes, RowCheck{
			Code:    fmt.Sprintf("T%d", i),
			Name:    fmt.Sprintf("ဝန်ထမ်းအမျိုးအစား %03d", i),
			Checked: true,
		})
	}

	return r
}
