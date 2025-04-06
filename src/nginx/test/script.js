
import http from 'k6/http';
import { check, sleep } from 'k6';

// Define the test configuration
export const options = {
    stages: [
        { duration: '10s', target: 20 }, // Ramp up to 30 users over 30 seconds
        { duration: '40s', target: 20 }, // Stay at 30 users for 1 minute
        { duration: '10s', target: 0 },  // Ramp down to 0 users over 30 seconds
    ],
};

// Main test function
export default function () {
    let data = {
        login: 'alivasilyev',
        password: '12345'
    };

    const url = 'http://localhost:9080/api/v2/sign-in';

    // Send a GET request
    const res = http.post(url, JSON.stringify(data), {
        headers: { 'Content-Type': 'application/json' },
    });

    // Check if the response is successful
    check(res, {
        'status is 200': (r) => r.status === 200,
        'response body is not empty': (r) => r.body.length > 0,
    });
}
