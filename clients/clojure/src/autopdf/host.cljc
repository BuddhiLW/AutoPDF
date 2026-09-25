(ns autopdf.host
  "The host file effects the client needs, one definition per host. JVM
   Clojure and ClojureWasm share clojure.java.io; clojurust names the same
   operations clojure.rust.io."
  (:require #?(:rust [clojure.rust.io :as io]
               :default [clojure.java.io :as io])))

(defn path
  "Join a directory and a file name."
  [dir file-name]
  (str dir "/" file-name))

(defn delete-file!
  "Remove the file at `p`; a missing file is not an error."
  [p]
  (io/delete-file p true))

(defn exists?
  "True when a file or directory exists at `p`."
  [p]
  #?(:rust (io/exists? p)
     :default (.exists (io/file p))))

(defn mkdirs!
  "Create directory `p` and any missing parents."
  [p]
  #?(:rust (io/make-parents (path p "x"))
     :default (.mkdirs (io/file p))))
