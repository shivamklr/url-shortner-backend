import http from "k6/http";
import { sleep, check } from "k6";
import { SharedArray } from "k6/data";

const writeUrls = new SharedArray("Write Data Json", function () {
    return JSON.parse(open("./sample-data/WriteData.json"));
});

const readUrls = new SharedArray("Read Data Json", function () {
    return JSON.parse(open("./sample-data/ReadData.json"));
});

export const options = {
    // discardResponseBodies: true,
    stages: [
        { duration: "10s", target: 3 }, // simulate ramp-up of traffic
        // { duration: "30s", target: 100 }, // simulate ramp-up of traffic
        // { duration: "2m", target: 200 }, // simulate peak traffic
        { duration: "1m", target: 0 }, // simulate ramp-down of traffic
    ],
    thresholds: {
        http_req_duration: ["p(95)<500"], // 95% of requests should be below 500ms
    },
};
const BASE_URL = "http://localhost:3000";

export default function () {
    const randomWriteIndex = Math.floor(Math.random() * writeUrls.length);
    const writeUrl = writeUrls[randomWriteIndex];

    writeRequest(writeUrl)
    
    const randomReadIndex = Math.floor(Math.random() * readUrls.length)
    const readUrl = readUrls[randomReadIndex]
    readRequest(readUrl)

    sleep(1);
}

export function writeRequest({original_url, expire_in}) {
    // write
    const payload = JSON.stringify({
        original_url: original_url,
        expire_in: expire_in,
    });
    const params = {
        headers: {
            "Content-Type": "application/json",
        },
    };
    const response = http.post(`${BASE_URL}/api/v1`, payload, params);
    check(response.status, { "status is 201": (r) => r === 201 });
}

export function readRequest({shorten_url}){
    const res = http.get(`${BASE_URL}/${shorten_url}`, { redirects: 0, name:"ReadUrlItem" });
    check(res.status, { "status is 302": (r) => r === 302 });
}
