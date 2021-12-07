package genstruct

import (
    "bytes"
    "fmt"
    "github.com/leeprince/gormstruct/internal/constants"
    "github.com/leeprince/gormstruct/internal/generate"
    "github.com/leeprince/gormstruct/internal/genfunc"
    "github.com/leeprince/gormstruct/internal/logger"
    "html/template"
)

/**
 * @Author: prince.lee <leeprince@foxmail.com>
 * @Date:   2021/11/9 下午11:49
 * @Desc:
 */

// ------------------- GenPackage
// 定义包名
func (p *GenPackage) SetPackage(pname string) {
    p.PackageName = pname
}

// 定义单个表的结构体。
func (p *GenPackage) SetStructs(sct GenStruct) {
    p.Struct = sct
}

// 定义单个表的结构体。
func (p *GenPackage) GenFileCtx() (strOut string) {
    var pa generate.PrintAtom
    
    // 定义包名：package
    pa.Add(constants.FilePackage, p.PackageName)
    logger.Infof("定义单个表的结构体。GenFileCtx.定义包名：package: %+v", pa.Generates())
    
    // 定义引入文件：import
    p.genImport()
    if len(p.Imports) > 0 {
        pa.Add(fmt.Sprintf("%s (", constants.FileImport))
        for _, i2 := range p.Imports {
            pa.Add(i2)
        }
        pa.Add(")")
    }
    logger.Infof("定义单个表的结构体。GenFileCtx.定义引入文件：import: %+v", pa.Generates())
    
    // 生成表的结构体
    if p.Struct.TableName != "" {
        for _, i2 := range p.Struct.Genrates() {
            pa.Add(i2)
        }
        logger.Infof("定义单个表的结构体。GenFileCtx.生成表的结构体: %+v", pa.Generates())
        
        // 添加表结构体对应的表名方法
        pa.Add(p.Struct.GenerateTableName())
        logger.Infof("定义单个表的结构体。GenFileCtx.添加表结构体对应的表名方法: %+v", pa.Generates())
        
    }
    
    // 添加获取表数据的方法
    for _, i2 := range p.FuncStrList {
        pa.Add(i2)
    }
    logger.Infof("定义单个表的结构体。GenFileCtx.添加获取表数据的基本方法: %+v", pa.Generates())
    
    // 拼接所有输出行的信息，并使用换行符换行
    logger.Infof("定义单个表的结构体。GenFileCtx.所有输出行的信息: %+v", pa.Generates())
    for _, v := range pa.Generates() {
        strOut += v + "\n"
    }
    logger.Infof("定义单个表的结构体。GenFileCtx.拼接所有输出行的信息，并使用换行符换行:\n%+v", strOut)
    
    return
}
func (p *GenPackage) genImport() {
    for _, i2 := range p.Struct.Elments {
        if v, ok := constants.ImportFile[i2.Type]; ok {
            p.AddImport(v)
        }
    }
}
func (p *GenPackage) AddImport(imp string) {
    p.Imports = append(p.Imports, imp)
}

// AddFuncStr add func coding string.添加函数串
func (p *GenPackage) AddFuncStr(src string) {
    p.FuncStrList = append(p.FuncStrList, src)
}

// ------------------- GenPackage ---end

// ----------------- GenStruct
func (s *GenStruct) SetStructName(structName string) {
    s.Name = structName
}
func (s *GenStruct) SetStructTableName(tabName string) {
    s.TableName = tabName
}
func (s *GenStruct) SetStructComment(Comment string) {
    s.Comment = Comment
}
func (s *GenStruct) SetStructElments(el ...GenElement) {
    s.Elments = append(s.Elments, el...)
}
func (s *GenStruct) Genrates() []string {
    var pa generate.PrintAtom
    pa.Add(fmt.Sprintf("// %s", s.Comment))
    pa.Add("type", s.Name, fmt.Sprintf("%s {", constants.FileStruct))
    for _, i2 := range s.Elments {
        pa.Add(i2.Genrates())
    }
    pa.Add("}")
    return pa.Generates()
}
func (s *GenStruct) GenerateTableName() string {
    templ, err := template.New("TEMP_GENTNF").Parse(genfunc.GetGenTableNameTemp())
    if err != nil {
        logger.Error("GenerateTableName err:", err)
    }
    data := struct {
        StructName string
        TableName  string
    }{
        StructName: s.Name,
        TableName:  s.TableName,
    }
    var buf bytes.Buffer
    templ.Execute(&buf, data)
    return buf.String()
}

// ----------------- GenStruct -- end

// ------------------- GenElement
func (e *GenElement) Genrates() string {
    var pa generate.PrintAtom
    if len(e.Comment) > 0 {
        pa.Add(e.Name, e.Type, e.TagString, fmt.Sprintf("// %s", e.Comment))
    } else {
        pa.Add(e.Name, e.Type, e.TagString)
    }
    return pa.Generates()[0]
}

// ------------------- GenElement -- end
