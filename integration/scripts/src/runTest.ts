import { ChainlinkNode } from './chainlinkNode'
import { fetchTests } from './tests'
import { fetchArgs } from './args'
import { fetchCredentials } from './common'
import * as assert from './asserts'

const main = async () => {
  const tests = fetchTests()

  const { chainlinkUrl } = fetchArgs()
  const credentials = fetchCredentials()
  const node = new ChainlinkNode(chainlinkUrl, credentials)

  const ctx = await assert.context(async (ctx) => {
    for (let i = 0; i < tests.length; i++) {
      const test = tests[i]
      await assert.newTest(test, async () => {
        const jobCount = (await node.getJobs()).meta?.count || 0
        let jobId: string
        await assert.it('creates job', ctx, async () => {
          jobId = await addJob(node, test.params)
          assert.isFalse(test, !jobId, 'got a job ID')
          const newJobCount = (await node.getJobs()).meta?.count
          assert.equals(test, newJobCount, jobCount + 1, 'job count should increase by 1')
        })

        await assert.it('runs job successfully', ctx, async () => {
          await assert.withRetry(async () => {
            const jobRuns = (await node.getJobRuns(jobId!)).meta?.count
            assert.equals(test, jobRuns, test.expectedRuns, 'job runs should increase')
          }, 30)

          await assert.withRetry(async () => {
            const jobRunStatus = (await node.getJobRuns(jobId!)).data[test.expectedRuns - 1]
              .attributes.status
            assert.equals(
              test,
              jobRunStatus,
              'completed',
              'last job run should be marked as completed',
            )
          }, 5)
        })
      })
    }
  })

  console.log()
  console.log('==== TEST RESULT ====')
  console.log('Tests passed:', ctx.successes)
  console.log('Tests failed:', ctx.fails)
  console.log('=====================')
  console.log()

  if (ctx.fails > 0) {
    process.exit(1)
  }
}

const addJob = async (node: ChainlinkNode, params: Record<string, any>): Promise<string> => {
  const jobspec = {
    initiators: [
      {
        type: 'external',
        params: {
          name: 'mock-client',
          body: params,
        },
      },
    ],
    tasks: [{ type: 'noop' }],
  }
  const Job = await node.createJob(jobspec)
  return Job.data.id!
}

main().then()
