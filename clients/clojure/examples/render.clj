;; Render examples/invoice.tex to a PDF by driving the unmodified autopdf binary.
;; The same file runs on every host:
;;
;;   cljw          -cp src examples/render.clj
;;   cljrs run     --src-path src examples/render.clj
;;   clojure       -M examples/render.clj
;;
;; AUTOPDF_BIN overrides the binary (default: autopdf on PATH).
(require '[autopdf.client :as autopdf]
         '[clojure.java.shell :refer [sh]]
         '[clojure.string :as str])

(def items [{:description "Consulting" :amount 1200}
            {:description "Support" :amount 300}])

(def result
  (autopdf/build {:template "invoice.tex"
                  :dir "examples"
                  :output "invoice.pdf"
                  :bin (or (System/getenv "AUTOPDF_BIN") "autopdf")
                  :vars {:number "2026-042"
                         :issuer "BuddhiLW"
                         :date "2026-09-25"
                         :client {:name "ACME Ltd." :address "1 Main St."}
                         :items items
                         :total (reduce + (map :amount items))}}))

;; A failure throws, which every host turns into a non-zero exit status.
(let [pdf (or (get-in result [:ok :pdf])
              (throw (ex-info "autopdf render failed" result)))
      ;; A PDF is binary; read only its ASCII magic and its size, through the
      ;; same process seam, so no host has to decode it.
      head (:out (sh "head" "-c" "5" pdf))
      size (first (str/split (str/trim (:out (sh "wc" "-c" pdf))) #"\s+"))]
  (when-not (= "%PDF-" head)
    (throw (ex-info "not a PDF" {:pdf pdf :head head})))
  (println "wrote" pdf (str "(" size " bytes, header " (pr-str head) ")")))
