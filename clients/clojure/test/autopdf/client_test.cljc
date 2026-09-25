(ns autopdf.client-test
  "autopdf.client against a recording stub runner: no binary, no LaTeX.
   Runs unchanged on JVM Clojure, ClojureWasm and clojurust."
  (:require [autopdf.client :as client]
            [autopdf.host :as host]
            [clojure.test :refer [deftest is testing]]))

(defrecord RecordingRunner [calls result]
  client/ProgramRunner
  (run-program [_ argv opts]
    (swap! calls conj {:argv argv
                       :opts opts
                       :config (slurp (host/path (:dir opts) (nth argv 3)))})
    result))

(defn- recording [result]
  (->RecordingRunner (atom []) result))

(def ^:private tmp
  (host/path (or (System/getenv "TMPDIR") "/tmp") "autopdf-client-test"))

(deftest json-encoding
  (is (= "{\"a\":[1,true,null,\"x\\\"y\\n\"],\"k\":\"v\"}"
         (client/->json {:a [1 true nil "x\"y\n"] :k :v}))))

(deftest plan-is-a-value
  (let [p (client/plan {:template "t.tex" :dir "d" :vars {:n 1}} "tok")]
    (is (= ["autopdf" "build" "t.tex" ".autopdf-tok.json" "clean"] (:argv p)))
    (is (= ".autopdf-tok.json" (:config-file p)))
    (is (= "d/output.pdf" (:pdf p)))
    (is (= (client/->json (client/config {:template "t.tex" :output "output.pdf" :vars {:n 1}}))
           (:config-json p))))
  (testing ":clean? false drops the flag; :bin flows through"
    (is (= ["/opt/autopdf" "build" "t.tex" ".autopdf-x.json"]
           (:argv (client/plan {:template "t.tex" :dir "d" :bin "/opt/autopdf" :clean? false} "x")))))
  (testing "a render without :template or :dir is refused"
    (is (thrown? Exception (client/plan {:dir "d"} "x")))))

(deftest execute-through-the-port
  (host/mkdirs! tmp)
  (testing "success: the runner sees the config on disk, which is removed afterwards"
    (let [runner (recording {:exit 0 :out "" :err ""})
          p (client/plan {:template "t.tex" :dir tmp :vars {:who "cljw"}} "ok")
          res (client/execute! runner p)
          [call] @(:calls runner)]
      (is (= {:ok {:pdf (:pdf p)}} res))
      (is (= (:argv p) (:argv call)))
      (is (= {:dir tmp} (:opts call)))
      (is (= (:config-json p) (:config call)))
      (is (not (host/exists? (host/path tmp (:config-file p)))))))
  (testing "a failed render is data, and the config is still removed"
    (let [p (client/plan {:template "t.tex" :dir tmp} "err")
          res (client/execute! (recording {:exit 1 :out "" :err "boom"}) p)]
      (is (= {:error {:exit 1 :err "boom"}} res))
      (is (not (host/exists? (host/path tmp (:config-file p))))))))
