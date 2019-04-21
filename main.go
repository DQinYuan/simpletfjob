package main

import (
	"bytes"
	"fmt"
	"github.com/dqinyuan/simpletfjob/kube"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
)

/*
模板提供的变量有：

 - .type
 - .index
 - .host
 - .TF_CONFIG
*/

const ARR_TEMPLATE = `{{range .}}"{{.}}:2222",{{end}}`

const TF_CONFIG_TEMPLATE =
`                {
                  "cluster": {
                    "ps": [{{.ps}}],
                    "worker": [{{.worker}}]
                  },
                  "task": {
                    "index": {{.index}},
                    "type": "{{.type}}"
                  }
                }
`

var tfTmpl *template.Template
var arrTmpl *template.Template

var bs  = new(bytes.Buffer)

func init() {
	tmpl, err := template.New("arr").Parse(ARR_TEMPLATE)
	if err != nil{
		log.Printf("template parse error, %v \n", err)
	}
	arrTmpl = tmpl


	tmpl, err = template.New("tfconfig").Parse(TF_CONFIG_TEMPLATE)
	if err != nil{
		log.Printf("template parse error, %v \n", err)
	}

	tfTmpl = tmpl
}

func formatArr(strArr []string)  string {
	bs.Reset()
	arrTmpl.Execute(bs, strArr)
	bs.Truncate(bs.Len() - 1)
	return bs.String()
}

func formatTtConfig(ps string, worker string, index int, tp string) string  {
	bs.Reset()
	tfTmpl.Execute(bs, map[string]string{
		"ps": ps,
		"worker": worker,
		"index": strconv.Itoa(index),
		"type": tp,
	})

	return bs.String()
}

func formatTmpl(tmpl *template.Template, params map[string]string) string {
	bs.Reset()
	tmpl.Execute(bs, params)
	return bs.String()
}

func decompose(nodes []string, psn int, psf string) (ps []string, worker []string) {
	if psf == ""{
		// psn 起作用, 选前n个node作为ps
		if (psn >= len(nodes)){
			log.Fatalln("psn can not bigger then nodes num")
		}
		return nodes[:psn], nodes[psn:]
	}

	// ToDo: 实现psf
	return nil, nil
}

var params = make(map[string]string)

func appendJob(tplat *template.Template, hosts []string, ps string, worker string, tp string, result *[]string) {
	for i, host := range hosts{
		params["host"] = host
		params["index"] = strconv.Itoa(i)
		params["type"] = tp
		params["TF_CONFIG"] = formatTtConfig(ps, worker, i, tp)
		*result = append(*result, formatTmpl(tplat, params))
	}
}

func transfer(tmplFile string, ps []string, worker []string) string {
	contentBytes, err := ioutil.ReadFile(tmplFile)
	if err != nil{
		log.Fatalf("read template file error, err %v\n", err)
	}
	content := strings.TrimSpace(string(contentBytes))
	jobTmpl, err := template.New("job").Parse(content)
	if err != nil{
		log.Fatalf("user template parse err, %v\n", err)
	}

	results := make([]string, 0, len(ps) + len(worker))
	psStr := formatArr(ps)
	workerStr := formatArr(worker)

	appendJob(jobTmpl, ps, psStr, workerStr, "ps", &results)
	appendJob(jobTmpl, worker, psStr, workerStr, "worker", &results)

	resultStr := strings.Join(results, "\n---\n")

	return resultStr
}

func resultFileName(originFileName string) string {
	suffix := ".yaml"
	if strings.HasSuffix(originFileName, suffix){
		originFileName = originFileName[:(len(originFileName) - 5)]
	}

	return fmt.Sprintf("%s_tfjob.yaml", originFileName)
}

func filterNodesByExec(nodes []string, excPath string) []string {
	bs, err := ioutil.ReadFile(excPath)
	if err != nil{
		log.Fatalf("read exec file %s fail, err %v\n", excPath, err)
	}

	excs := strings.Split(strings.TrimSpace(string(bs)), "\n")
	excSet := make(map[string]bool)
	for _, exc := range excs{
		excSet[strings.TrimSpace(exc)] = true
	}

	filtered := make([]string, 0)
	for _, node := range nodes{
		if !excSet[node]{
			filtered = append(filtered, node)
		}
	}

	return filtered
}

func filterNodes(nodes []string, exc bool, num int) []string {
	if exc{
		nodes = filterNodesByExec(nodes, "exc")
	}

	if num != -1 && num != 0{
		nodes = nodes[:num]
	}

	return nodes
}

var (
	num int
	exc bool
	psn int
	psf string
)

func cmdFun(cmd *cobra.Command, args []string) {
	tmplFile := args[0]
	nodes := kube.ClusterNodenames()
	nodes = filterNodes(nodes, exc, num)

	ps, worker := decompose(nodes, psn, psf)

	transfered := transfer(tmplFile, ps, worker)

	fullPath, err := filepath.Abs(resultFileName(tmplFile))
	if err != nil{
		log.Fatalf("result file path error, %v\n", err)
	}
	ioutil.WriteFile(fullPath, []byte(transfered), 0644)
}

func main() {
	rootCmd := &cobra.Command{
		Use: "simpletfjob user_template",
		Short: "transfer tensorflow job template to k8s job yaml",
		Args:cobra.MinimumNArgs(1),
		Run: cmdFun,
	}

	rootCmd.Flags().IntVar(&psn, "psn", 1,
		"parameter server num")
	rootCmd.Flags().StringVar(&psf, "psf", "",
		"special parameter server by file, not yet implemented")
	rootCmd.Flags().IntVarP(&num, "num", "N", -1,
		"server number you want to use")
	rootCmd.Flags().BoolVarP(&exc, "exc", "E", false,
		"the file name, in which are recorded server names you want to exclude each line")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(rootCmd.UsageString())
		os.Exit(1)
	}
}
