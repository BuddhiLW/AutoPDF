(ns autopdf.client-test
  "autopdf.client against a recording stub runner and, for the stdin seam, a
   real child process (cat). No autopdf binary, no LaTeX. Runs unchanged on
   JVM Clojure, ClojureWasm and clojurust."
  (:require [autopdf.client :as client]
            [clojure.test :refer [deftest is testing]]))

(defrecord RecordingRunner [calls result]
  client/ProgramRunner
  (run-program [_ argv opts]
    (swap! calls conj {:argv argv :opts opts})
    result))

(defn- recording [result]
  (->RecordingRunner (atom []) result))

(deftest json-encoding
  (is (= "{\"a\":[1,true,null,\"x\\\"y\\n\"],\"k\":\"v\"}"
         (client/->json {:a [1 true nil "x\"y\n"] :k :v}))))

(deftest plan-is-a-value
  (let [p (client/plan {:template "t.tex" :dir "d" :vars {:n 1}})]
    (is (= ["autopdf" "build" "t.tex" "-" "clean"] (:argv p)))
    (is (= "d/output.pdf" (:pdf p)))
    (is (= (client/->json (client/config {:template "t.tex" :output "output.pdf" :vars {:n 1}}))
           (:config-json p)))
    (is (not (contains? p :config-file)) "no config file is planned"))
  (testing ":clean? false drops the flag; :bin flows through"
    (is (= ["/opt/autopdf" "build" "t.tex" "-"]
           (:argv (client/plan {:template "t.tex" :dir "d" :bin "/opt/autopdf" :clean? false})))))
  (testing "a render without :template or :dir is refused"
    (is (thrown? Exception (client/plan {:dir "d"})))))

(deftest execute-through-the-port
  (testing "success: the runner gets the config as its standard input"
    (let [runner (recording {:exit 0 :out "" :err ""})
          p (client/plan {:template "t.tex" :dir "d" :vars {:who "cljw"}})
          res (client/execute! runner p)
          [call] @(:calls runner)]
      (is (= {:ok {:pdf (:pdf p)}} res))
      (is (= (:argv p) (:argv call)))
      (is (= {:dir "d" :in (:config-json p)} (:opts call)))))
  (testing "a failed render is data"
    (let [p (client/plan {:template "t.tex" :dir "d"})
          res (client/execute! (recording {:exit 1 :out "" :err "boom"}) p)]
      (is (= {:error {:exit 1 :err "boom"}} res)))))

(deftest shell-runner-writes-stdin
  (testing "the default runner hands :in to the child and closes it"
    (is (= {:exit 0 :out "{\"a\":1}" :err ""}
           (client/run-program client/shell-runner ["cat"] {:dir "." :in "{\"a\":1}"}))))
  (testing "without :in the child still runs"
    (is (zero? (:exit (client/run-program client/shell-runner ["true"] {:dir "."}))))))
