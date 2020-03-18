const url = 'https://us-west-2-1.aws.cloud2.influxdata.com'
const orgId = '275ac1e8a61d71f2'
const bucketName = 'performance'
const authToken =
  'iEoLyuknDOIzw7ByDlwAmTwPADgFD0C7Jni6u3UMc5iL18Ph4V6bnQex753GBv_L3yqQCMOI1yr7sJBPsYChtg=='

export type AppPerformance = any

const INTERVAL = 120

export const performanceLineWriter = function performanceLineWriter(
  appPerformance: AppPerformance
) {
  fetch(`${url}/api/v2/write?org=${orgId}&bucket=${bucketName}&precision=ms`, {
    method: 'POST',
    mode: 'cors',
    headers: {
      Authorization: `Token ${authToken}`,
    },
    body: `performance,environment=dev,interval=${INTERVAL} fps=${
      appPerformance.currentFPS
    },maxFPS=${appPerformance.maxFPS},totalFrames=${
      appPerformance.totalFrames
    },averageFPS=${appPerformance.averageFPS},timeInCurrentFrameMS=${
      appPerformance.timeInCurrentFrame
    },maxTimeInFrameMS=${appPerformance.maxTimeInFrame} ${Date.now()}`,
  })
}
