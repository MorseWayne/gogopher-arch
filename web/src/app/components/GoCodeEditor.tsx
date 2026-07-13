import { useMemo } from "react";
import CodeMirror from "@uiw/react-codemirror";
import { autocompletion, snippetCompletion, type Completion, type CompletionContext, type CompletionResult } from "@codemirror/autocomplete";
import { go } from "@codemirror/lang-go";
import { oneDark } from "@codemirror/theme-one-dark";
import { EditorView } from "@codemirror/view";

export function GoCodeEditor({
  value,
  onChange,
  concepts = [],
  height = "28rem",
  readOnly = false,
  ariaLabel = "Go code editor",
  syntax = "go",
}: {
  value: string;
  onChange: (value: string) => void;
  concepts?: string[];
  height?: string;
  readOnly?: boolean;
  ariaLabel?: string;
  syntax?: "go" | "plain";
}) {
  const conceptKey = concepts.join("|");
  const extensions = useMemo(
    () => [
      ...(syntax === "go" ? [go(), autocompletion({ override: [createGoCompletionSource(concepts)] })] : []),
      EditorView.contentAttributes.of({ "aria-label": ariaLabel }),
    ],
    [ariaLabel, conceptKey, syntax],
  );

  return (
    <div className="overflow-hidden rounded-2xl border border-slate-800 bg-black shadow-inner">
      <CodeMirror
        aria-label={ariaLabel}
        value={value}
        height={height}
        theme={oneDark}
        editable={!readOnly}
        readOnly={readOnly}
        basicSetup={{
          lineNumbers: true,
          foldGutter: true,
          highlightActiveLine: true,
          highlightSelectionMatches: true,
          bracketMatching: true,
          closeBrackets: true,
          autocompletion: true,
        }}
        extensions={extensions}
        onChange={(nextValue) => {
          if (!readOnly) onChange(nextValue)
        }}
        className="text-sm leading-6"
      />
    </div>
  );
}

type CourseCompletion = Completion & {
  concepts?: string[];
};

const goKeywordCompletions: CourseCompletion[] = [
  { label: "package", type: "keyword", detail: "声明包名", apply: "package main" },
  { label: "import", type: "keyword", detail: "导入包", apply: "import \"fmt\"" },
  { label: "func", type: "keyword", detail: "函数声明", apply: "func" },
  { label: "type", type: "keyword", detail: "类型声明", apply: "type" },
  { label: "struct", type: "keyword", detail: "结构体", apply: "struct" },
  { label: "interface", type: "keyword", detail: "接口", apply: "interface" },
  { label: "map", type: "keyword", detail: "哈希表", apply: "map" },
  { label: "range", type: "keyword", detail: "遍历", apply: "range" },
  { label: "go", type: "keyword", detail: "启动 goroutine", apply: "go" },
  { label: "defer", type: "keyword", detail: "延迟执行", apply: "defer" },
  { label: "select", type: "keyword", detail: "channel 多路复用", apply: "select" },
  { label: "if err != nil", type: "keyword", detail: "错误处理", apply: "if err != nil {\n\treturn err\n}" },
];

const goSnippetCompletions: CourseCompletion[] = [
  snippet("func main() {\n\t${}\n}", { label: "func main", type: "function", detail: "main 入口函数" }),
  snippet("fmt.Println(${})", { label: "fmt.Println", type: "function", detail: "打印一行输出" }),
  snippet("for ${i} := 0; ${i} < ${n}; ${i}++ {\n\t${}\n}", { label: "for i := 0", type: "keyword", detail: "计数循环" }),
  snippet("for ${_, v} := range ${items} {\n\t${}\n}", { label: "for range", type: "keyword", detail: "range 遍历" }),
  snippet("if ${condition} {\n\t${}\n}", { label: "if", type: "keyword", detail: "条件分支" }),
  snippet("if ${value}, ok := ${m}[${key}]; ok {\n\t${}\n}", { label: "map ok", type: "keyword", detail: "map 存在性检查" }, ["map"]),
  snippet("${counts} := map[string]int{}", { label: "map[string]int{}", type: "variable", detail: "字符串计数 map" }, ["map"]),
  snippet("sort.Strings(${keys})", { label: "sort.Strings", type: "function", detail: "稳定排序字符串切片" }, ["sort", "deterministic output", "stable output"]),
  snippet("var ${builder} strings.Builder\n${builder}.WriteString(${text})", { label: "strings.Builder", type: "class", detail: "高效构造字符串" }, ["strings.Builder"]),
  snippet("type ${Name} struct {\n\t${Field} ${string}\n}", { label: "type struct", type: "class", detail: "声明结构体" }, ["struct"]),
  snippet("json.Marshal(${value})", { label: "json.Marshal", type: "function", detail: "编码 JSON" }, ["json", "DTO", "data boundary"]),
  snippet("tests := []struct {\n\tname string\n\tinput string\n\twant string\n}{\n\t{\n\t\tname: ${\"case\"},\n\t\tinput: ${\"input\"},\n\t\twant: ${\"want\"},\n\t},\n}", { label: "table-driven tests", type: "text", detail: "表驱动测试用例" }, ["table-driven tests", "test cases"]),
  snippet("for _, tt := range tests {\n\tt.Run(tt.name, func(t *testing.T) {\n\t\t${}\n\t})\n}", { label: "t.Run", type: "function", detail: "子测试循环" }, ["table-driven tests", "test cases"]),
  snippet("if got != want {\n\tt.Fatalf(\"got %v, want %v\", got, want)\n}", { label: "t.Fatalf", type: "function", detail: "测试失败信息" }, ["failure messages", "regression test"]),
  snippet("strings.TrimSpace(${value})", { label: "strings.TrimSpace", type: "function", detail: "去除首尾空白" }, ["strings"]),
  snippet("strings.Fields(${value})", { label: "strings.Fields", type: "function", detail: "按连续空白拆分" }, ["strings", "edge cases"]),
];

function createGoCompletionSource(concepts: string[]) {
  const options = getCompletionOptions(concepts);

  return (context: CompletionContext): CompletionResult | null => {
    const word = context.matchBefore(/[\w.]*/);

    if (!word || (word.from === word.to && !context.explicit)) {
      return null;
    }

    return {
      from: word.from,
      options,
      validFor: /^[\w.]*$/,
    };
  };
}

function getCompletionOptions(concepts: string[]) {
  if (concepts.length === 0) {
    return [...goKeywordCompletions, ...goSnippetCompletions];
  }

  const normalizedConcepts = concepts.map((concept) => concept.toLowerCase());
  const conceptMatches = goSnippetCompletions.filter((completion) => completion.concepts?.some((concept) => normalizedConcepts.includes(concept.toLowerCase())));

  return [...goKeywordCompletions, ...conceptMatches, ...goSnippetCompletions.filter((completion) => !completion.concepts)];
}

function snippet(template: string, completion: Completion, concepts?: string[]): CourseCompletion {
  return {
    ...snippetCompletion(template, completion),
    concepts,
  };
}
