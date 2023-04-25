import http, { BatchRequest, BatchRequests, ObjectBatchRequest } from "k6/http";
import { sleep } from "k6";
import { SharedArray } from "k6/data";

const writeUrls = new SharedArray("Write Data Json", function () {
    return JSON.parse(open("../../sample-data/WriteData.json"));
});


interface WriteDataObject {
    original_url: string,
    expire_in: number
}

export const options = {
    discardResponseBodies: true,
    stages: [
        { duration: "10s", target: 3 }, // simulate ramp-up of traffic
        { duration: "2m", target: 100 }, // simulate ramp-up of traffic
        // { duration: "2m", target: 200 }, // simulate peak traffic
        { duration: "1m", target: 0 }, // simulate ramp-down of traffic
    ],
    thresholds: {
        http_req_duration: ["p(95)<500"], // 95% of requests should be below 500ms
    },
};
const BASE_URL = "http://localhost:3000";

export default function () {

    const writeBatch = new Array<BatchRequest>();
    for (let i = 0; i < 50; i++) {
        const randomWriteIndex = Math.floor(Math.random() * writeUrls.length);
        const writeUrl = writeUrls[randomWriteIndex] as WriteDataObject;
        const payload = JSON.stringify({
            original_url: writeUrl.original_url,
            expire_in: writeUrl.expire_in,
        });

        const writeRequest: ObjectBatchRequest = {
            method: 'POST',
            url: `${BASE_URL}/api/v1`,
            params: {
                headers: {
                    "Content-Type": "application/json",
                },
                tags: { name: "WriteUrlItem" },
            },
            body: payload
        }
        writeBatch.push(writeRequest)
    }
    RequestBatch(writeBatch);

    sleep(1);
}

export function writeRequest({ original_url, expire_in }: WriteDataObject) {
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
    http.post(`${BASE_URL}/api/v1`, payload, params);
}

export function RequestBatch(Requests: BatchRequests) {
    http.batch(Requests)
}
