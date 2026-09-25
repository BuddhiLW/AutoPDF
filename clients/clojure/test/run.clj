;; Portable test entry point: the same file runs on every host.
;;
;;   cljw          -cp src:test test/run.clj
;;   cljrs run     --src-path src --src-path test test/run.clj
;;   clojure       -M:test test/run.clj
;;
;; A failure throws, which every host turns into a non-zero exit status.
(require 'autopdf.client-test)

(let [{:keys [fail error]} (clojure.test/run-tests 'autopdf.client-test)]
  (when (pos? (+ fail error))
    (throw (ex-info "autopdf.client-test failed" {:fail fail :error error}))))
