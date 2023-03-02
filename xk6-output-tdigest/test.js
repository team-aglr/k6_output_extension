import http from 'k6/http';
import { sleep } from 'k6';

export default function () {
  http.get('https://test-api.k6.io');
  http.get('https://test-api.k6.io');
  http.get('https://test-api.k6.io');
  sleep(12);
}


function nearestNSecondsLowerBound(seconds, d=new Date()) {
  let ms = 1000 * seconds; // convert number of seconds to ms
  return Math.round(d / ms) * ms;
}