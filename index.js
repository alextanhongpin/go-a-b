'use strict'
const ctx = document.getElementById('canvas').getContext('2d')
const updateButton = document.getElementById('update')
const simulateButton = document.getElementById('simulate')
const arm1 = document.getElementById('arm1')
const arm2 = document.getElementById('arm2')
const arm3 = document.getElementById('arm3')

const bandit = new EpsilonGreedy({
  epsilon: 0.1,
  n: 3
})

const chart = new Chart(ctx, {
  type: 'bar',
  data: {
    labels: ['a', 'b', 'c'],
    datasets: [
      {
        label: 'Reward',
        backgroundColor: 'rgb(0, 99, 132)',
        borderColor: 'rgb(255, 99, 132)',
        data: [0, 0, 0]
      },
      {
        label: 'Pulls',
        backgroundColor: 'rgb(255, 99, 0)',
        borderColor: 'rgb(255, 99, 132)',
        data: [0, 0, 0]
      }
    ]
  },
  options: {}
})

let cumulativeReward = []
let rewards = []
let count = 0
function updateChart (arm) {
  const reward = Math.random() < 0.5 ? 0 : 1
  bandit.update(arm, reward)

  rewards[count] = reward
  cumulativeReward[count] = rewards[count - 1] ? rewards[count - 1] + reward : 0
  const rewardDataset = chart.data.datasets[0].data
  const pullsDataset = chart.data.datasets[1].data
  // Randomly assign the reward
  rewardDataset[arm] += reward
  // Add the pull count
  pullsDataset[arm] += 1

  console.log('\ncount', count)
  console.log('arm', arm)
  console.log('pulls', JSON.stringify(pullsDataset))
  console.log('reward', JSON.stringify(rewardDataset))
  const totalReward = rewardDataset.reduce((a, b) => a + b, 0)
  const totalPulls = pullsDataset.reduce((a, b) => a + b, 0)
  console.log('reward / pull', JSON.stringify(totalReward / totalPulls))

  chart.update()
  count += 1
}

updateButton.addEventListener('click', (evt) => {
  const arm = bandit.selectArm()
  updateChart(arm)
}, false)

simulateButton.addEventListener('click', (evt) => {
  Array(100).fill(0).forEach(() => {
    const arm = bandit.selectArm()
    updateChart(arm)
  })
}, false)

arm1.addEventListener('click', (evt) => {
  updateChart(0)
}, false)

arm2.addEventListener('click', (evt) => {
  updateChart(1)
}, false)

arm3.addEventListener('click', (evt) => {
  updateChart(2)
}, false)
