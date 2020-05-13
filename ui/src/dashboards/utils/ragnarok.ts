

export const listServices = async () => {
  let servicesResponse

  try {
    servicesResponse = await fetch('http://localhost:8081/api/services')
  } catch (err) {
    throw err
  }

  if (!servicesResponse.ok) {
    console.error(servicesResponse)
    return []
  }

  return await servicesResponse.json()
}

export const getInstance = async (instance) => {
  let instanceResponse
  try {
     instanceResponse = await fetch('/ragnarok/instances', {
      method: 'POST',
      body: JSON.stringify(instance),
      headers: {
        'Content-Type': 'application/json',
      },
    })
  } catch (error) {
    console.error(error)
  }

  if (!instanceResponse.ok) {
    console.error(instanceResponse)
    return null
  }

  return await instanceResponse.json()
}

export const startForecasting = async (instanceId, inputQuery) => {
  const body = {
    instanceId,
    operationName: 'Forecast',
    inputQuery,
    outputDatabase: 'ds-bucket',
    outputMeasurement: 'forecasting',
    params: {Days: '365'},
  }

  let activitiesResponse
  try {
     activitiesResponse = await fetch('/ragnarok/activities', {
      method: 'POST',
      body: JSON.stringify(body),
      headers: {
        'Content-Type': 'application/json',
      },
    })
  } catch (error) {
    console.error(error)
  }

  if (!activitiesResponse.ok) {
    console.error(activitiesResponse)
    return
  }

  const {activityId} = await activitiesResponse.json()
  return activityId
}

export const runWhenComplete = async (activityId, callback) => {
  let activitiesResponse = await fetch('/ragnarok/activities', {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
  })

  let activities = await activitiesResponse.json()
  let targetActivity = activities.find(activity => activity.activityId === activityId)

  if (!targetActivity) {
    throw new Error(`Couldn't find any activities with id ${activityId}`)
  }

  // task is complete, run our callback function
  if (targetActivity.status === 'Completed') {
    callback()
    return
  }

  // Task is still running, check again in 1 second
  setTimeout(() => {
    runWhenComplete(activityId, callback)
  }, 1000)

}
