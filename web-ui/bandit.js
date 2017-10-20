class EpsilonGreedy {
  constructor ({ epsilon, n }) {
    this.epsilon = epsilon // The default epsilon
    this.n = n // Number of arms
    this.counts = Array(n).fill(0)
    this.values = Array(n).fill(0)
  }

  selectArm () {
    const isExploiting = Math.random() > this.epsilon
    if (isExploiting) { // Exploit!
      // Index of the best arm
      return this.values.indexOf(Math.max(...this.values))
    } else { // Explore!
      // Index of randomly selected arms
      return Math.floor(Math.random() * this.n)
    }
  }

  update (chosenArm, reward) { // index of pulled arm
    const n = ++this.counts[chosenArm]
    const v = this.values[chosenArm]
    this.values[chosenArm] = (v * (n - 1) + reward) / n
  }
}

class BernoulliArm {
  constructor (p) {
    // Probability of getting a reward
    this.p = p
  }
  pull () {
    return Math.random() > this.p ? 0 : 1
  }
}

function categoricalDraw (probs) {
  const z = Math.random()
  let cumulativeProb = 0
  for (let i = 0; i < probs.length; i += 1) {
    const prob = probs[i]
    cumulativeProb += prob
    if (cumulativeProb > z) {
      return i
    }
  }
  return probs.length - 1
}

function simulate (AlgoClass, options, arms, horizon) {
  const chosenArm = Array(horizon).fill()
  const rewards = Array(horizon).fill()
  const cumulativeRewards = Array(horizon).fill()

  const algo = new AlgoClass(options)
  let cumulativeReward = 0

  for (let t = 0; t < horizon; t += 1) {
    const i = algo.selectArm()
    const arm = arms[i]
    const reward = arm.pull()
    algo.update(i, reward)

    chosenArm[t] = i
    rewards[t] = reward
    cumulativeReward += reward
    cumulativeRewards[t] = cumulativeReward
  }
  return { chosenArm, rewards, cumulativeRewards, algo }
}

class Softmax {
  constructor ({ temperature, n }) { // temperature is equal to epsilon
    this.temperature = temperature
    this.counts = Array(n).fill(0)
    this.values = Array(n).fill(0)
  }

  selectArm () {
    const z = this.values.map((v) => Math.exp(v / this.temperature)).reduce((a, b) => a + b, 0)
    const probs = this.values.map((v) => Math.exp(v / this.temperature) / z)
    return categoricalDraw(probs)
  }

  update (chosenArm, reward) {
    const n = ++this.counts[chosenArm]
    const value = this.values[chosenArm]
    const newValue = ((n - 1) / n) * value + (1 / n) * reward
    this.values[chosenArm] = newValue
  }
}

function sum (arr) {
  return arr.reduce((a, b) => a + b, 0)
}

class AnnealingSoftmax {
  constructor ({ n }) {
    this.counts = Array(n).fill(0)
    this.values = Array(n).fill(0)
  }

  selectArm () {
    const t = sum(this.counts) + 1
    const temperature = 1 / Math.log(t + 0.0000001)
    const z = sum(this.values.map((v) => Math.exp(v / temperature)))
    const probs = this.values.map((v) => Math.exp(v / temperature) / z)
    return categoricalDraw(probs)
  }

  update (chosenArm, reward) {
    const n = ++this.counts[chosenArm]
    const value = this.values[chosenArm]
    const newValue = ((n - 1) / n) * value + (1 / n) * reward
    this.values[chosenArm] = newValue
  }
}

// The Upper-Confidence Bound Algorithm
class UCB1 {
  constructor ({ n }) {
    this.counts = Array(n).fill(0)
    this.values = Array(n).fill(0)
  }
  selectArm () {
    const nArms = this.counts.length
    // Prevent cold-start by ensure all the choices has been represented at least once
    const zeroArm = this.counts.findIndex(v => v === 0)
    if (zeroArm !== -1) {
      return zeroArm
    }

    const totalCounts = this.counts.reduce((a, b) => a + b, 0)
    const ucbValues = Array(nArms).fill(0).map((_, arm) => {
      const count = this.counts[arm]
      const value = this.values[arm]
      const bonus = Math.sqrt((2 * Math.log(totalCounts)) / count)
      return bonus + value
    })
    return ucbValues.indexOf(Math.max(...ucbValues))
  }

  update (chosenArm, reward) {
    this.counts[chosenArm] = this.counts[chosenArm] + 1
    const n = this.counts[chosenArm]
    const value = this.values[chosenArm]
    const newValue = ((n - 1) / n) * value + (1 / n) * reward
    this.values[chosenArm] = newValue
  }
}

function testEpsilonGreedy () {
  const means = [0.1, 0.1, 0.1, 0.1, 0.9]
  const nArms = means.length
  const horizon = 100000
  const arms = means.map(v => new BernoulliArm(v))
  const results = simulate(EpsilonGreedy, { epsilon: 0.1, n: nArms }, arms, horizon)

  console.log(arms)
  console.log(results.chosenArm[99999])
  console.log(results.cumulativeRewards[99999])

  // maxReward = horizon * maxP = 100000 * 0.1 = 10000
  // regret = maxReward - totalReward = 10000 - 8906 = 1094
  // maxReward * epsilon = 10000 * 0.1 = 1000 = 1094 = regret
}

function testSoftmax () {
  const means = [0.1, 0.1, 0.1, 0.1, 0.9]
  const nArms = means.length
  const horizon = 100000
  const arms = means.map((mean) => new BernoulliArm(mean))
  const results = simulate(Softmax, { temperature: 0.1, n: nArms }, arms, horizon)
  console.log(results.chosenArm[horizon - 1])
  console.log(results.cumulativeRewards[horizon - 1])
}

function testAnnealingSoftmax () {
  const means = [0.1, 0.1, 0.1, 0.1, 0.9]
  const nArms = means.length
  const horizon = 1000
  const arms = means.map((mean) => new BernoulliArm(mean))
  const results = simulate(AnnealingSoftmax, { n: nArms }, arms, horizon)
  console.log(results.chosenArm[horizon - 1])
  console.log(results.cumulativeRewards[horizon - 1])
}

function testUCB () {
  const means = [0.1, 0.1, 0.1, 0.1, 0.9]
  const nArms = means.length
  const horizon = 1000
  const arms = means.map((mean) => new BernoulliArm(mean))
  const results = simulate(UCB1, { n: nArms }, arms, horizon)
  console.log(results.chosenArm[horizon - 1])
  console.log(results.cumulativeRewards[horizon - 1])
}

testUCB()
