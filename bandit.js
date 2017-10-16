class EpsilonGreedy {
  constructor ({ epsilon, n }) {
    this.epsilon = epsilon
    this.n = n // Number of arms
    this.counts = Array(n).fill(0)
    this.values = Array(n).fill(0)
  }

  selectArm () {
    if (Math.random() > this.epsilon) { // Exploit!
      // Index of the best arm
      return this.values.indexOf(Math.max(...this.values))
    } else { // Explore!
      // Index of randomly selected arms
      return Math.floor(Math.random() * this.n)
    }
  }
  update (i, reward) { // i - index of pulled arm
    const n = ++this.counts[i]
    const v = this.values[i]
    this.values[i] = (v * (n - 1) + reward) / n
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
  return { chosenArm, rewards, cumulativeRewards }
}

function main () {
  const nArms = 5
  const horizon = 100000
  const arms = Array(nArms).fill().map(_ => new BernoulliArm(Math.random() / 10))
  const results = simulate(EpsilonGreedy, { epsilon: 0.1, n: nArms }, arms, horizon)

  console.log(arms)
  console.log(results.cumulativeRewards[99999])

  // maxReward = horizon * maxP = 100000 * 0.1 = 10000
  // regret = maxReward - totalReward = 10000 - 8906 = 1094
  // maxReward * epsilon = 10000 * 0.1 = 1000 = 1094 = regret
}
