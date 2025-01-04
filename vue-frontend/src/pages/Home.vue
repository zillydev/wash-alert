<script setup lang="ts">
import { ref, inject, type Ref } from 'vue'
import BookPopup from '../components/BookPopup.vue';
import { send } from '../utils/websocket';

type Machine = {
  IsMachine: boolean
  MachineNumber: number
  Status: string
  TimeBooked: string | null
  HoursBooked: number
  MinutesBooked: number
  Timer: string
}

type Floor = {
  FloorName: string
  Rows: number
  Columns: number
  Grid: Machine[][]
}

const socket = inject('socket') as WebSocket
const token = inject('token') as Ref<string, string>
const loaded = ref(false)
const isPopupVisible = ref(false)
const floorNumber = ref(0)
const rowNumber = ref(0)
const colNumber = ref(0)
const displayError = ref(false)
const errorMessage = ref('Please enter valid values')
const hours = ref(0)
const minutes = ref(0)
const booked = ref(localStorage.getItem('booked') === 'true')
var storageBookedFloor = localStorage.getItem('bookedFloor')
var storageBookedRow = localStorage.getItem('bookedRow')
var storageBookedCol = localStorage.getItem('bookedCol')
const bookedFloor = ref(storageBookedFloor !== null ? Number(storageBookedFloor) : null)
const bookedRow = ref(storageBookedRow !== null ? Number(storageBookedRow) : null)
const bookedCol = ref(storageBookedCol !== null ? Number(storageBookedCol) : null)
var floors = ref([] as Floor[])

function calculateTimeLeft(hours: number, minutes: number, bookedTime: string | null) {
  if (!bookedTime) {
    return ''
  }
  const startTime = new Date(bookedTime)
  const endTime = new Date(startTime.getTime() +
    (hours * 60 * 60 * 1000) +
    (minutes * 60 * 1000))
  const now = new Date()

  if (now > endTime) return 'Done'
  if (now < startTime) return 'Not started'

  const diff = endTime.getTime() - now.getTime()

  const hoursLeft = Math.floor(diff / (1000 * 60 * 60))
  const minutesLeft = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60))
  const secondsLeft = Math.floor((diff % (1000 * 60)) / 1000)

  return `${hoursLeft.toString().padStart(2, '0')}:${minutesLeft.toString().padStart(2, '0')}:${secondsLeft.toString().padStart(2, '0')}`
}

type Timer = {
  floorNumber: number
  rowNumber: number
  colNumber: number
  hours: number
  minutes: number
  bookedTime: string | null
}

var timers: Timer[] = []

// function startTimer(floorNumber: number, rowNumber: number, colNumber: number, hours: number, minutes: number, bookedTime: string | null) {
//   if (!bookedTime) {
//     return ''
//   }
//   setInterval(() => {
//     const timeLeft = calculateTimeLeft(hours, minutes, bookedTime)
//     floors.value[floorNumber].Grid[rowNumber][colNumber].Timer = timeLeft
//   }, 1000)
// }
var interval: number | undefined
function startTimers() {
  interval = setInterval(() => {
    for (var i = 0; i < timers.length; i++) {
      const timeLeft = calculateTimeLeft(timers[i].hours, timers[i].minutes, timers[i].bookedTime)
      floors.value[timers[i].floorNumber].Grid[timers[i].rowNumber][timers[i].colNumber].Timer = timeLeft
    }
  }, 1000)
}

socket.onmessage = (event) => {
  var data = JSON.parse(event.data)
  var type = data.Type
  if (type === 'data') {
    let message = data.Message as Floor[]
    console.log(message)
    for (var i = 0; i < message.length; i++) {
      for (var j = 0; j < message[i].Grid.length; j++) {
        for (var k = 0; k < message[i].Grid[j].length; k++) {
          if (message[i].Grid[j][k].Status === 'Booked') {
            // startTimer(i, j, k, message[i].Grid[j][k].HoursBooked, message[i].Grid[j][k].MinutesBooked, message[i].Grid[j][k].TimeBooked)
            timers.push({
              floorNumber: i,
              rowNumber: j,
              colNumber: k,
              hours: message[i].Grid[j][k].HoursBooked,
              minutes: message[i].Grid[j][k].MinutesBooked,
              bookedTime: message[i].Grid[j][k].TimeBooked
            })
            if (timers.length === 1) {
              startTimers()
            }
          }
        }
      }
    }
    floors.value = message
    loaded.value = true
  } else if (type === 'booked') {
    let message = data.Message
    console.log(message)

    floors.value[message.FloorNumber].Grid[message.RowNumber][message.ColumnNumber].Status = 'Booked'
    floors.value[message.FloorNumber].Grid[message.RowNumber][message.ColumnNumber].TimeBooked = message.TimeBooked
    floors.value[message.FloorNumber].Grid[message.RowNumber][message.ColumnNumber].HoursBooked = message.HoursBooked
    floors.value[message.FloorNumber].Grid[message.RowNumber][message.ColumnNumber].MinutesBooked = message.MinutesBooked
    floors.value[message.FloorNumber].Grid[message.RowNumber][message.ColumnNumber].Timer = 'Booking...'

    timers.push({
      floorNumber: message.FloorNumber,
      rowNumber: message.RowNumber,
      colNumber: message.ColumnNumber,
      hours: message.Hours,
      minutes: message.Minutes,
      bookedTime: message.TimeBooked
    })
    if (timers.length === 1) {
      startTimers()
    }
  } else if (type === 'done') {
    let message = data.Message
    console.log(message)
    // floors.value[message.FloorNumber].Grid[message.RowNumber][message.ColumnNumber].Booked = false
    floors.value[message.FloorNumber].Grid[message.RowNumber][message.ColumnNumber].Status = 'Done'
    floors.value[message.FloorNumber].Grid[message.RowNumber][message.ColumnNumber].Timer = ''
    floors.value[message.FloorNumber].Grid[message.RowNumber][message.ColumnNumber].TimeBooked = null
    floors.value[message.FloorNumber].Grid[message.RowNumber][message.ColumnNumber].HoursBooked = 0
    floors.value[message.FloorNumber].Grid[message.RowNumber][message.ColumnNumber].MinutesBooked = 0

    timers = timers.filter(timer => timer.floorNumber !== message.FloorNumber || timer.rowNumber !== message.RowNumber || timer.colNumber !== message.ColumnNumber)
    if (timers.length === 0) {
      clearInterval(interval)
    }
  } else if (type === 'collected') {
    let message = data.Message
    console.log(message)
    // floors.value[message.FloorNumber].Grid[message.RowNumber][message.ColumnNumber].Booked = false
    floors.value[message.FloorNumber].Grid[message.RowNumber][message.ColumnNumber].Status = 'Empty'
    floors.value[message.FloorNumber].Grid[message.RowNumber][message.ColumnNumber].Timer = 'Booking...'
    floors.value[message.FloorNumber].Grid[message.RowNumber][message.ColumnNumber].TimeBooked = null
    floors.value[message.FloorNumber].Grid[message.RowNumber][message.ColumnNumber].HoursBooked = 0
    floors.value[message.FloorNumber].Grid[message.RowNumber][message.ColumnNumber].MinutesBooked = 0

    if (message.FloorNumber === bookedFloor.value && message.RowNumber === bookedRow.value && message.ColumnNumber === bookedCol.value) {
      localStorage.setItem('booked', 'false')
      localStorage.setItem('bookedFloor', '')
      localStorage.setItem('bookedRow', '')
      localStorage.setItem('bookedCol', '')
      booked.value = false
      bookedFloor.value = null
      bookedRow.value = null
      bookedCol.value = null
    }
  }
}

function openPopup(floorIndex: number, rowIndex: number, colIndex: number) {
  if (booked.value || !floors.value[floorIndex].Grid[rowIndex][colIndex].IsMachine || floors.value[floorIndex].Grid[rowIndex][colIndex].Status === 'Booked' || floors.value[floorIndex].Grid[rowIndex][colIndex].Status === 'Done' || floors.value[floorIndex].Grid[rowIndex][colIndex].Status === 'Not Functional') {
    return
  }
  isPopupVisible.value = true
  floorNumber.value = floorIndex
  rowNumber.value = rowIndex
  colNumber.value = colIndex
}

function closePopup() {
  isPopupVisible.value = false
  displayError.value = false
  floorNumber.value = 0
  rowNumber.value = 0
  colNumber.value = 0
  hours.value = 0
  minutes.value = 0
}

function validate() {
  if (hours.value < 0 || minutes.value < 0 || (hours.value === 0 && minutes.value === 0) || hours.value > 2 || minutes.value > 59) {
    displayError.value = true
    return
  }
  var bookedTime = new Date().toISOString()
  send(socket, 'book', { "FloorNumber": floorNumber.value, "RowNumber": rowNumber.value, "ColumnNumber": colNumber.value, "Hours": hours.value, "Minutes": minutes.value, "TimeBooked": bookedTime, "Token": token.value })

  localStorage.setItem('booked', 'true')
  localStorage.setItem('bookedFloor', floorNumber.value.toString())
  localStorage.setItem('bookedRow', rowNumber.value.toString())
  localStorage.setItem('bookedCol', colNumber.value.toString())
  booked.value = true
  bookedFloor.value = floorNumber.value
  bookedRow.value = rowNumber.value
  bookedCol.value = colNumber.value

  closePopup()
}

function collect(floorNumber: number, rowNumber: number, colNumber: number) {
  console.log("Collected", floorNumber, rowNumber, colNumber)
  localStorage.setItem('booked', 'false')
  localStorage.setItem('bookedFloor', '')
  localStorage.setItem('bookedRow', '')
  localStorage.setItem('bookedCol', '')
  booked.value = false
  bookedFloor.value = null
  bookedRow.value = null
  bookedCol.value = null
  send(socket, 'collect', { "FloorNumber": floorNumber, "RowNumber": rowNumber, "ColumnNumber": colNumber })
}
</script>

<template>
  <BookPopup :isVisible="isPopupVisible" @update:isVisible="closePopup()">
    <div class="flex flex-col gap-10 justify-center items-center">
      <h1 class="text-3xl font-bold underline text-white">Book {{ floors[floorNumber].FloorName }}, Machine {{
        floors[floorNumber].Grid[rowNumber][colNumber].MachineNumber }}</h1>
      <div class="flex gap-5">
        Set timer:
        <div class="flex gap-2">
          <input type="number" min="0" max="59" v-model="hours">
          hr,
          <input type="number" min="0" max="59" v-model="minutes">
          min
        </div>
      </div>
      <div v-if="displayError" class="text-red-500">{{ errorMessage }}</div>
      <button class="bg-gray-500 px-4 py-2 rounded-md" @click="validate()">Book</button>
    </div>
  </BookPopup>
  <div v-if="loaded" class="w-full flex justify-center">
    <div class="container flex flex-col gap-3 p-3">
      <div v-for="(floor, floorIndex) in floors" :key="floorIndex" class="flex flex-col mb-20">
        <h1 class="text-3xl font-bold underline mb-5 text-center">
          {{ floor.FloorName }}
        </h1>

        <div class="overflow-x-auto max-w-full self-center">
          <div v-for="(row, rowIndex) in floor.Grid" :key="rowIndex" :class="`grid gap-5 mb-5 min-w-max`" :style="{
            gridTemplateColumns: `repeat(${floor.Columns}, minmax(0, 1fr))`,
          }">
            <template v-for="(machine, colIndex) in row">
              <div v-if="machine.IsMachine" :key="machine.MachineNumber"
                class="flex flex-col justify-center items-center w-[100px] @container" :style="{
                  backgroundColor: (machine.Status !== 'Booked' && machine.Status !== 'Not Functional') ? 'green' : 'red',
                }" @click="openPopup(floorIndex, rowIndex, colIndex)">
                <div class="text-[20cqi] text-center">{{ machine.Status !== 'Booked' ? machine.Status : machine.Timer }}</div>
                <img src="/machine-icon2.webp" alt="machine icon" />
                <div class="text-[25cqi]">{{ machine.MachineNumber }}</div>
                <button v-if="machine.Status === 'Done' && bookedFloor === floorIndex && bookedRow === rowIndex && bookedCol === colIndex"
                  class="bg-gray-500 px-4 py-2 rounded-md"
                  @click.stop="collect(floorIndex, rowIndex, colIndex)">Collect</button>
              </div>
              <div v-else class="w-[100x]"></div>
            </template>
          </div>
        </div>
      </div>
    </div>
  </div>
  <div v-else>Loading...</div>
</template>
