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
        let res = http.get("http://myapp.local/configs");
        check(res, {
            "status is 200": (r) => r.status === 200,
        });
    });

    // GET request
    group("GET", function() {
        let res = http.get("http://myapp.local/configs/pod-1");
        check(res, {
            "status is 200": (r) => r.status === 200,
        });
    });

    // POST request
    //group("POST", function() {
        //let res = http.post("http://myapp.local/configs", { verb: "post" }, { headers: { "Content-Type": "application/json"   }});
        //check(res, {
            //"status is 201": (r) => r.status === 200,
        //});
    //});

    // PUT request
    //group("PUT", function() {
        //let res = http.put("http://myapp.local/configs", JSON.stringify({ verb: "put" }), { headers: { "Content-Type": "application/json" }});
        //check(res, {
            //"status is 200": (r) => r.status === 200,
        //});
    //});

    // PATCH request
    //group("PATCH", function() {
        //let res = http.patch("http://myapp.local/configs", JSON.stringify({ verb: "patch" }), { headers: { "Content-Type": "application/json" }});
        //check(res, {
            //"status is 200": (r) => r.status === 200,
        //});
    //});

    // DELETE request
    //group("DELETE", function() {
        //let res = http.del("http://myapp.local/configs/pod-1");
        //check(res, {
            //"status is 200": (r) => r.status === 200,
        //});
    //});
};
