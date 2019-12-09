import http from "k6/http";
import { group, check  } from "k6";


export let options = {
    thresholds: {
        'http_req_duration{kind:html}': ["avg<=10"],
        'http_reqs': ["rate>100"],
    }
};

export default function() {
    // GET request
    group("GET", function() {
        let res = http.get("http://192.168.99.101:30010/configs");
        check(res, {
            "status is 200": (r) => r.status === 200,
            "is verb correct": (r) => r.json().args.verb === "get",
        });
    });

    // GET request
    group("GET", function() {
        let res = http.get("http://192.168.99.101:30010/configs/pod-1");
        check(res, {
            "status is 200": (r) => r.status === 200,
            "is verb correct": (r) => r.json().args.verb === "get",
        });
    });

    // POST request
    group("POST", function() {
        let res = http.post("http://192.168.99.101:30010/configs", { verb: "post" }, { headers: { "Content-Type": "application/json"   }});
        check(res, {
            "status is 200": (r) => r.status === 200,
            "is verb correct": (r) => r.json().form.verb === "post",
        });
    });

    // PUT request
    group("PUT", function() {
        let res = http.put("http://192.168.99.101:30010/configs", JSON.stringify({ verb: "put" }), { headers: { "Content-Type": "application/json" }});
        check(res, {
            "status is 200": (r) => r.status === 200,
            "is verb correct": (r) => r.json().json.verb === "put",
        });
    });

    // PATCH request
    group("PATCH", function() {
        let res = http.patch("http://192.168.99.101:30010/configs", JSON.stringify({ verb: "patch" }), { headers: { "Content-Type": "application/json" }});
        check(res, {
            "status is 200": (r) => r.status === 200,
            "is verb correct": (r) => r.json().json.verb === "patch",
        });
    });

    // DELETE request
    group("DELETE", function() {
        let res = http.del("http://192.168.99.101:30010/configs/pod-1");
        check(res, {
            "status is 200": (r) => r.status === 200,
            "is verb correct": (r) => r.json().args.verb === "delete",
        });
    });
};
