package glweb

import (
	"encoding/base64"
	"io"
	"net/http"
	"sort"
	"strings"

	//	"glweb/glweb/totp"
)

func unauthorized(res http.ResponseWriter) {
	res.Header().Set("WWW-Authenticate", "Basic realm=\""+BasicRealm+"\"")
	http.Error(res, "Not Authorized", http.StatusUnauthorized)
}

func auth(res http.ResponseWriter, req *http.Request) bool {
	auth := req.Header.Get("Authorization")
	if len(auth) < 6 || auth[:6] != "Basic " {
		unauthorized(res)
		return false
	}
	b, err := base64.StdEncoding.DecodeString(auth[6:])
	if err != nil {
		unauthorized(res)
		return false
	}
	tokens := strings.SplitN(string(b), ":", 2)
	if len(tokens) != 2 || !authfn(tokens[0], tokens[1]) {
		unauthorized(res)
		return false
	}
	return true
}
func authfn(u, p string) bool {
	return u == p //totp.Authenticate([]byte(MD5(u)), p, totpo)
}

//func htasks(glf map[string]*glf, w http.ResponseWriter, r *http.Request) {
//	tasksstr := ""

//	glrs := make(map[string]*Task)

//	ks := make([]string, 0)
//	for _, g := range glf {
//		for n, t := range g.tasks {
//			glrs[n] = t
//			ks = append(ks, n)
//		}
//	}
//	sort.Strings(ks)
//	ts := make([]*Task, 0)
//	for _, k := range ks {
//		ts = append(ts, glrs[k])
//	}
//	for _, t := range ts {
//		tasksstr = tasksstr +
//			`<li>
//           	 <a href="#" mce_href="#"> <span class="title">` + t.Name + `</span></a>
//              <table id="l1">
//                <tr>
//                  <td class="intro mytd">Version:</td>
//                  <td>` + t.Version + `</td>
//                </tr>
//                <tr>
//                  <td class="intro mytd">Action:</td>
//                  <td>` + t.Action + `</td>
//                </tr>
//                <tr>
//                  <td class="intro mytd">Method:</td>
//                  <td>` + t.Method + `</td>
//                </tr>
//                <tr>
//                  <td class="intro mytd">Info:</td>
//                  <td>
//                  <div class="info">` + t.Info + `</div></td>
//                </tr>
//              </table>
//           </li>`
//	}
//	nr := strings.Replace(taskshtml, "{{tasks}}", tasksstr, -1)
//	io.WriteString(w, nr)
//}
func htasks(glfs map[string]*glf, glps map[string]string,
	w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	if r.FormValue("task") != "" {
		if g, ok := glfs[glps[r.FormValue("task")]]; ok {
			ks := make([]string, 0)
			for n, _ := range g.tasks {
				ks = append(ks, n)
			}
			sort.Strings(ks)

			tasksstr := ""
			for _, k := range ks {
				t := g.tasks[k]
				tasksstr = tasksstr +
					`<li>
			           	 <a href="#" mce_href="#"> <span class="title">` + t.Name + `</span></a>
			              <table id="l1">
			                <tr>
			                  <td class="intro mytd">Version:</td>
			                  <td>` + t.Version + `</td>
			                </tr>
			                <tr>
			                  <td class="intro mytd">Action:</td>
			                  <td>` + t.Action + `</td>
			                </tr>
			                <tr>
			                  <td class="intro mytd">Method:</td>
			                  <td>` + t.Method + `</td>
			                </tr>
			                <tr>
			                  <td class="intro mytd">Info:</td>
			                  <td>
			                  <div class="info">` + t.Info + `</div></td>
			                </tr>
			              </table>
			           </li>`
			}
			resp := strings.Replace(taskshtml, "{{tasks}}", tasksstr, -1)
			io.WriteString(w, strings.Replace(resp, "{{taskname}}", "-"+g.name, -1))
		} else {
			http.NotFound(w, r)
		}

	} else {
		ks := make([]string, 0)
		for k, _ := range glps {
			ks = append(ks, k)
		}
		sort.Strings(ks)
		tasksstr := ""
		for _, k := range ks {
			tasksstr = tasksstr + `<li><a href="/?task=` + k + `"> <span class="title">` + k + `</span></a></li>`
		}
		resp := strings.Replace(taskshtml, "{{tasks}}", tasksstr, -1)
		io.WriteString(w, strings.Replace(resp, "{{taskname}}", "", -1))
	}
}

const taskshtml = `<!Doctype html><html><head><meta http-equiv=Content-Type content="text/html;charset=utf-8"><link rel="alternate" type="text/xml" href="/" /><style type="text/css">
		BODY { color: #000000; background-color: white; font-family: Verdana; margin-left: 0px; margin-top: 0px; }
		#content { margin-left: 30px; font-size: .70em; padding-bottom: 2em; }
		P { color: #000000; margin-top: 0px; margin-bottom: 12px; font-family: Verdana; }
		ul { margin-top: 10px; margin-left: 20px; }
		li { margin-top: 10px; color: #000000; }
    table {display: none;}
		.heading1 { color: #ffffff; font-family: Tahoma; font-size: 26px; font-weight: normal; background-color: #6A5ACD; margin-top: 0px; margin-bottom: 0px; margin-left: -30px; padding-top: 10px; padding-bottom: 3px; padding-left: 15px; width: 105%; }
	.glwb {text-decoration: none;color: white;}
    .title {font-weight:bold; margin-left: -5px; font-size:20px;color:#6A5ACD;}
		.intro { font-weight:bold; }
    .info {width:640px;}
    .mytd  {display: block; float:right; clear:both; padding:0px; margin:0px;}
    </style>
    <script src="http://cdn.bootcss.com/jquery/2.1.4/jquery.min.js"></script>
  <script type="text/javascript">
    $(document).ready(function() {
        var as = $("a");
        as.click(function() {
            var aNode = $(this);
            var lis = aNode .nextAll("table");
            lis.toggle("show");
        });
    });
  </script>
  <title>GLWEB Tasks 服务</title>
</head><body><div id="content"><p class="heading1"><a href="/#" class="glwb">GLWEB</a>{{taskname}}</p><br><span><ul><li>符号：&:和  |:或</li><li>GET参数:使用URL参数方式。例如：/list?start=1&per=10</li><li>POST参数:使用POST参数方式</li><li>分页参数(start:开始页&per:每页条数):必须同时使用。</li><li>时间字段：使用两个界限时间比对。例如：time=2015-01-01/2015-01-02</li>{{tasks}}</ul></span></div></body></html>
`
