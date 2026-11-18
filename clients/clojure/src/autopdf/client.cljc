(ns autopdf.client
  "Render AutoPDF templates from Clojure by running the autopdf binary.

   No bindings and no rebuild: the seam is the CLI. A render is planned as a
   value (the argv, the JSON config, the PDF path), then executed through a
   ProgramRunner port. The config reaches autopdf on standard input
   (`autopdf build TEMPLATE -`), so nothing is written beside the template.

   Portable across JVM Clojure, ClojureWasm (cljw) and clojurust (cljrs):
   each provides clojure.java.shell/sh with :in.

   Strata: ports (ProgramRunner) / promote (config, ->json, plan: pure) /
   boundary (execute!: the only effect, one process) / facade (build)."
  (:require [clojure.java.shell :as shell]
            [clojure.string :as str]))

;; ── Port ────────────────────────────────────────────────────────────────────

(defprotocol ProgramRunner
  (run-program [this argv opts]
    "Run argv with opts {:dir d :in s}, writing s to the program's standard
     input and closing it; return {:exit n :out s :err s}. A non-zero exit is
     data, not an exception."))

(defrecord ShellRunner []
  ProgramRunner
  (run-program [_ argv {:keys [dir in]}]
    (apply shell/sh (concat argv (cond-> [:dir dir] in (conj :in in))))))

(def shell-runner
  "The default runner: clojure.java.shell/sh."
  (->ShellRunner))

;; ── Promote: pure ───────────────────────────────────────────────────────────

(defn- json-str [s]
  (str "\""
       (str/escape s {\" "\\\"" \\ "\\\\" \newline "\\n" \return "\\r" \tab "\\t"
                      \backspace "\\b" \formfeed "\\f"})
       "\""))

(defn ->json
  "Encode maps, sequentials, strings, keywords, numbers, booleans and nil as JSON."
  [x]
  (cond
    (nil? x) "null"
    (boolean? x) (str x)
    (number? x) (str x)
    (string? x) (json-str x)
    (keyword? x) (json-str (name x))
    (symbol? x) (json-str (str x))
    (map? x) (str "{" (str/join "," (map (fn [[k v]] (str (->json (if (keyword? k) (name k) (str k))) ":" (->json v))) x)) "}")
    (sequential? x) (str "[" (str/join "," (map ->json x)) "]")
    :else (json-str (str x))))

(defn config
  "The autopdf config map for one render."
  [{:keys [template output engine vars]}]
  {:template template
   :output output
   :engine (or engine "pdflatex")
   :variables (or vars {})
   :conversion {:enabled false :formats []}})

(defn plan
  "The render described by opts, as a value: what to run, the JSON config to
   send on its standard input, and where the PDF lands."
  [{:keys [template dir output bin clean?]
    :or {output "output.pdf" bin "autopdf" clean? true}
    :as opts}]
  (when-not (and template dir)
    (throw (ex-info "autopdf.client: a render needs :template and :dir" {:opts opts})))
  {:dir dir
   :config-json (->json (config (assoc opts :output output)))
   :argv (cond-> [bin "build" template "-"] clean? (conj "clean"))
   :pdf (str dir "/" output)})

;; ── Boundary: the only effects ──────────────────────────────────────────────

(defn execute!
  "Run the plan through `runner`, sending its config on standard input.
   Returns {:ok {:pdf path}} or {:error {:exit n :err stderr}}."
  [runner {:keys [dir config-json argv pdf]}]
  (let [{:keys [exit err]} (run-program runner argv {:dir dir :in config-json})]
    (if (zero? exit)
      {:ok {:pdf pdf}}
      {:error {:exit exit :err err}})))

;; ── Facade ──────────────────────────────────────────────────────────────────

(defn build
  "Render `template` with `vars` to a PDF by running the autopdf binary.

   opts:
     :template  template file, relative to :dir (required)
     :dir       working directory holding the template (required)
     :vars      variables map; nested maps and vectors are fine
     :output    PDF file name written into :dir (default \"output.pdf\")
     :engine    LaTeX engine (default \"pdflatex\")
     :bin       autopdf executable, a name on PATH or a path (default \"autopdf\")
     :clean?    remove LaTeX auxiliary files afterwards (default true)
     :runner    a ProgramRunner (default shell-runner)

   Returns {:ok {:pdf path}} or {:error {:exit n :err stderr}}. A failed
   render is data; only a missing :template or :dir throws."
  [{:keys [runner] :or {runner shell-runner} :as opts}]
  (execute! runner (plan opts)))
