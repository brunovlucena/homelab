import http from "k6/http";
import { group, check  } from "k6";

export let options = {
    vus: 30,
    thresholds: {
        'http_req_duration{kind:html}': ["avg<=10"],
        'http_reqs': ["rate>100"],
        "http_req_duration": ["p(95)<500"],
        "check_failure_rate": [
            // Global failure rate should be less than 1%
            "rate<0.01",
            // Abort the test early if it climbs over 5%
            { threshold: "rate<=0.05", abortOnFail: true },
        ],
    },
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
        let res = http.get("http://myapp.local/configs/pod-2");
        check(res, {
            "status is 200": (r) => r.status === 200,
        });
    });

    //// POST request
    //let payload = '{"metadata":{"limits":{"cpu":{"enabled":true,"value":"512m"},"memory":{"enabled":false,"value":"1024Mi"}},"monitoring":{"enabled":false}},"name":"pod-2boooo"}';
    ////let body = JSON.stringify(payload);
    //group("POST", function() {
        //let res = http.post("http://myapp.local/configs", { verb: "post" }, payload, { headers: { "Content-Type": "application/json" }});
        //// Use JSON.parse to deserialize the JSON (instead of using the r.json() method)
        //let j = JSON.parse(res.body);
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
    group("DELETE", function() {
        let res = http.del("http://myapp.local/configs/pod-1-idonotexist");
        check(res, {
            "status is 422": (r) => r.status === 422,
        });
    });
};
