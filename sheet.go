package beep

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Sheet - music sheet
type Sheet struct {
	ID       int
	Name     string
	Dir      string
	Notation string
	URL      string
	file     *os.File
}

// NewSheet returns new music sheet
func NewSheet(name, dir, notation, url string) *Sheet {
	s := &Sheet{
		Name:     name,
		Dir:      dir,
		URL:      url,
		Notation: notation,
	}
	return s
}

// Path returns music sheet path
func (s *Sheet) Path() string {
	var filename string
	var name = s.Name
	if s.ID > 0 {
		name = fmt.Sprintf("%d-%s", s.ID, s.Name)
	}
	filename = filepath.Join(HomeDir(), "sheets", s.Dir, name)
	return filename
}

// Save persists music sheet
func (s *Sheet) Save() error {
	dir := filepath.Join(HomeDir(), "sheets", s.Dir)
	os.MkdirAll(dir, 0755)
	opt := os.O_CREATE | os.O_TRUNC | os.O_WRONLY
	file, err := os.OpenFile(s.Path(), opt, 0644)
	if err != nil {
		return err
	}
	s.file = file
	defer s.file.Close()
	fmt.Fprint(s.file, s.Notation)
	return nil
}

// Load reads music sheet from file
func (s *Sheet) Load() error {
	sheetID := stringNumber(strings.Split(s.Name, "-")[0])
	for _, sheet := range BuiltinMusic {
		if sheet.ID == sheetID {
			s.ID = sheet.ID
			s.Dir = sheet.Dir
			s.URL = sheet.URL
			s.Notation = sheet.Notation
			return nil
		}
	}
	buf, err := ioutil.ReadFile(s.Path())
	if err != nil {
		return err
	}
	s.Notation = string(buf)
	return nil
}

// Exists checks existing music sheet file
func (s *Sheet) Exists() bool {
	filename := filepath.Join(HomeDir(), "sheets", s.Dir, s.Name)
	_, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return true
}

// Delete removes music sheet file
func (s *Sheet) Delete() error {
	err := os.Remove(s.Path())
	if err != nil {
		return err
	}
	return nil
}

func sheetSearch(keyword string) []string {
	var names []string
	keyword = strings.ToLower(keyword)
	root := filepath.Join(HomeDir(), "sheets") + string(os.PathSeparator)
	for _, sheet := range BuiltinMusic {
		name := strings.TrimPrefix(sheet.Path(), root)
		if len(keyword) == 0 || strings.Contains(strings.ToLower(name), keyword) {
			names = append(names, name)
		}
	}
	walkFn := func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			name := strings.TrimPrefix(path, root)
			if len(keyword) == 0 || strings.Contains(strings.ToLower(name), keyword) {
				names = append(names, name)
			}
		}
		return nil
	}
	sort.Strings(names)
	filepath.Walk(root, walkFn)
	return names
}

// BuiltinMusic stores built-in music scores
var BuiltinMusic = []*Sheet{
	{
		ID:   1,
		Name: "mozart-k33b-klavierstuck-in-f.txt",
		Dir:  "beep",
		URL:  "http://imslp.org/images/1/15/TN-Mozart%2C_Wofgang_Amadeus-NMA_09_27_Band_02_I_01_KV_33b.jpg",
		Notation: `# Mozart K33b
VP SA8 SR9
A9HRDE cc DScszs|DEc DQzDE[|cc DScszs|DEc DQz DE[|vv DSvcsc|DEvs ]v|cc DScszs|VN
A3HLDE [n z,    |cHRq HLz, |[n z,    |cHRq HLz,  |sl z,    |]m   pb|z, ]m    |

A9HRDE cz [c|ss DSsz]z|DEs] ps|DSsz][ z][p|DEpDQ[ [|VN
A3HLDE [n ov|]m [n    |  pb ic|  n,   lHRq|HLnc DQ[|

A9HRDE cc DScszs|DEc DQzDE[|cc DScszs|DEc DQz DE[|vv DSvcsc|DEvs ]v|cc DScszs|VN
A3HLDE [n z,    |cHRq HLz, |[n z,    |cHRq HLz,  |sl z,    |]m   pb|z, ]m    |

A9HRDE cz [c|ss DSsz]z|DEs] ps|DSsz][ z][p|DEpDQ[ [|VN
A3HLDE [n ov|]m [n    |  pb ic|  n,   lHRq|HLnc DQ[|

A9HRDS DERE] DS][p[ |][p[ ][p[  |DE] DQp DEi|REc DScszs|cszs |cszs|DEcDQzDE[|REv DSvcsc|DEvs ]v|VN
A3HLDE DEcHRq HLvHRw|HLbHRe HLvw|cHRq   HLic|[n  ]m    |z,   |]m  |zn   z,  |sl  [,    |z. DQp |

A9HRDE REc DScszs|DEcz [c|REs DSsz]z|DEs] ps|DSsz][ z][p|DE[DSitDQr|VN
A3HLDE z,  ]m    |[n   ov|]m  [n    |pb   ic|nz     sc  |DQn      [|

A9HRDS DERE] DS][p[ |][p[ ][p[  |DE] DQp DEi|REc DScszs|cszs |cszs|DEcDQzDE[|REv DSvcsc|DEvs ]v|VN
A3HLDE DEcHRq HLvHRw|HLbHRe HLvw|cHRq   HLic|[n  ]m    |z,   |]m  |zn   z,  |sl  [,    |z. DQp |

A9HRDE REc DScszs|DEcz [c|REs DSsz]z|DEs] ps|DSsz][ z][p|DE[DSitDQrRQ|VN
A3HLDE z,  ]m    |[n   ov|]m  [n    |pb   ic|nz     sc  |DQn      [RQ|
`,
	},
	{
		ID:   2,
		Name: "passacaglia-handel-halvorsen.txt",
		Dir:  "beep",
		URL:  "https://azmusicfest.org/app/uploads/Passacaglia-Handel-Halvorsen-Pianistos-2.pdf",
		Notation: `# Passacaglia - Handel Halvorsen
VP T5 SA9 SD9 SS9 SR9
# DQ - 130
A6HRDE RERERERE RERERERE|RERERERE RERERERE|icxc   zc]c  |[cpc   ocic    |VN
A4HLDE z,HReq   yqeq    |HLz,HReq yqeq    |HLz,HReq yqeq|HLov,n HRwHLn,n|
# 5
A6HRDE uxzx     ]x[x        |pxox ixux    |yz]z      [zpz        |oziz uzyz    |DQz DEa= DQDDa DEz|VN
A4HLDE pmHRwHL. HRtHL.HRwHL.|icmb HRqHLbmb|[nmHRqHL, HRrHL,HRqHL,|ov,[ HRwHLn,n|pb  .k   HReHLk .k|
# 10
A6HRDW z          |DEcioi   pi[i|]izi   xici    |xuiu     oupu        |[u]u zuxu    |VN
A4HLDE z,HReq yqeq|HLz,HReq yqeq|HLov,n HRwHLn,n|]mHRwHL. HRtHL.HRwHL.|icmb HRqHLbmb|
# 15
A6HRDE zyuy     iyoy        |py[y ]yzy    |DQz DEa= DQDDa DEz|DWz        |DEcH7qHR.H7q  HR,H7qHRmH7q|VN
A4HLDE pbHRqHL, HRrHL,HRqHL,|ov,n HRwHLn,n|pb  .k   HReHLk .k|z,HReq yqeq|z,HReq yqeq               |
# 20   8
A6HRDE nH7qHRbH7q HRvH7qHRcH7q|HRx.,.   m.n.        |b.v. c.x.    |z,m,      n,b,        |v,c, x,z,    |VN
A4HLDE ov,n       HRwHLn,n    |pmHRwHL. HRtHL.HRwHL.|icmb HRqHLbmb|[nmHRqHL, HRrHL,HRqHL,|ov,[ HRwHLn,n|
# 25   8
A6HRDQ , DEkj DQDDk DE,|DW,        |DEH7qHRcvc bcnc|mc,c   .cH7qHRc|.xcx     vxbx        |VN
A4HLDE pb.k   HReHLk .k|z,HReq yqeq|HLz,HReq   yqeq|HLov,n HRwHLn,n|]mHRwHL. HRtHL.HRwHL.|
# 30   8
A6HRDE nxmx ,x.x    |.zxz     czvz        |bznz mz,z    |DQ, DEkj DQDDk DE,|DW,        |VN
A4HLDE icmb HRqHLbmb|pbHRqHL, HRrHL,HRqHL,|ov,n HRwHLn,n|pb  .k   HReHLk .k|z,HReq yqeq|
# 35
A6HRDE iici   xizi|]i[i   pioi    |uuxu     zu]u        |[upu ouiu    |yyzy     ]y[y        |VN
A4HLDE z,HReq yqeq|HLov,n HRwHLn,n|pbHRqHL, HRrHL,HRqHL,|icmb HRqHLbmb|pbHRqHL, HRrHL,HRqHL,|
# 40                                               # 8va---------------------------
A6HRDE pyoy iyuy    |DQy DE65 DQDD6 DEy|DWy        |DEcH7qHRx. z,]m|[npb   ovic    |VN
A4HLDE ov,n HRwHLn,n|pb.k     HReHLk .k|z,HReq yqeq|HLz,HReq   yqeq|HLov,n HRwHLn,n|
# 45   8
A6HRDE x.z,   ]m[n|pbov   icux    |yzt]     r[ep        |woqi uHL.HRyHL,|HRDQy DE65 DQDD6 DEy|VN
A4HLDE z,HReq yqeq|HLov,n HRwHLn,n|]mHRwHL. HRtHL.HRwHL.|ob,n ,n,n      |pb.k       HReHLk .k|
# 50
A6HRDW y          |DEicxc   zc]c|[cpc   ocic    |uxzx     ]x[x        |pxox ixux    |VN
A4HLDE z,HReq yqeq|HLz,HReq yqeq|HLov,n HRwHLn,n|pmHRwHL. HRtHL.HRwHL.|icmb HRqHLbmb|
# 55
A6HRDE yz]z      [zpz        |oziz uzyz    |DQz DEa= DQDDa DEz|DWz        |DEcioi   pi[i|VN
A4HLDE [nmHRqHL, HRrHL,HRqHL,|ov,[ HRwHLn,n|pb  .k   HReHLk .k|z,HReq yqeq|HLz,HReq yqeq|
# 60
A6HRDE ]izi xici    |xuiu     oupu        |[u]u zuxu    |zyuy     iyoy        |py[y ]yzy    |VN
A4HLDE ov,n HRwHLn,n|]mHRwHL. HRtHL.HRwHL.|icmb HRqHLbmb|pbHRqHL, HRrHL,HRqHL,|ov,n HRwHLn,n|
# 65
A6HRDQ z DEa= DQDDa DEz  |DWz        |DEicxc zc]c|[cpc ocic|uxzx  ]x[x|VN
A4HLDE pb  .k   HReHLk .k|z,HReq yqeq|HLDWC2,z   |C2vo     |C2m]      |
# 70
T3
A6HRDE pxox ixux|yz]z [zpz|oziz uzyz|DQz DEa= DQDDa DEz|DWz |VN
A4HLDW C2ci     |C2n[     |C2vo     |C2bp              |C2,z|
`,
	},
}
