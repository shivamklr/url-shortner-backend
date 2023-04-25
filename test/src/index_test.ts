import http, { BatchRequest, BatchRequests, ObjectBatchRequest } from "k6/http";
import { sleep } from "k6";
import { SharedArray } from "k6/data";

const writeUrls = new SharedArray("Write Data Json", function () {
    return JSON.parse(open("../../sample-data/WriteData.json"));
});

const readUrls = new SharedArray("Read Data Json", function () {
    return JSON.parse(open("../../sample-data/ReadData.json"));
});


interface ReadDataObject {
    original_url: string,
    shorten: string
}

interface WriteDataObject {
    original_url: string,
    expire_in: number
}

export const options = {
    discardResponseBodies: true,
    stages: [
        { duration: "10s", target: 3 }, // simulate ramp-up of traffic
        { duration: "3m", target: 100 }, // simulate ramp-up of traffic
        // { duration: "2m", target: 200 }, // simulate peak traffic
        { duration: "2m", target: 0 }, // simulate ramp-down of traffic
    ],
    thresholds: {
        http_req_duration: ["p(95)<500"], // 95% of requests should be below 500ms
    },
};

const BASE_URL = `${__ENV.MY_HOSTNAME}`;
export default function () {
    const randomWriteIndex = Math.floor(Math.random() * writeUrls.length);
    const writeUrl = writeUrls[randomWriteIndex] as WriteDataObject;

    writeRequest(writeUrl);

    const readBatch = new Array<BatchRequest>();
    for (let i = 0; i < 1000; i++) {
        const randomReadIndex = Math.floor(Math.random() * readUrls.length);
        const readUrl = readUrls[randomReadIndex] as ReadDataObject;
        const readRequest: ObjectBatchRequest = {
            method: 'GET',
            url: `${BASE_URL}/${readUrl.shorten}`,
            params: {
                tags: { name: "ReadUrlItem" },
                redirects: 0
            }
        }
        readBatch.push(readRequest)
    }
    readRequestBatch(readBatch);

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

export function readRequestBatch(readRequests: BatchRequests) {
    http.batch(readRequests)
}
