import assert from 'node:assert/strict';
import test from 'node:test';

import { STALE_REVIEW_WINDOW_HOURS, triageReviewRoom } from './review-room-triage.js';

const NOW = Date.parse('2026-03-11T12:30:00Z');

function item(overrides) {
  return {
    text: 'review item',
    kind: 'runtime_signal',
    severity: 'medium',
    source: 'agent_job.failed',
    created_at: '2026-03-11T12:00:00Z',
    updated_at: '2026-03-11T12:00:00Z',
    ...overrides,
  };
}

test('triageReviewRoom separates blockers, residue, and strategy', () => {
  const triage = triageReviewRoom({
    open: [
      item({
        text: 'fresh failed agent job',
        source: 'agent_job.failed',
        updated_at: '2026-03-11T11:10:00Z',
      }),
      item({
        text: 'old failed verifier',
        source: 'verification.failed',
        updated_at: '2026-03-10T01:10:00Z',
      }),
      item({
        text: 'hold merge until telegram boundary is hardened',
        source: 'codex_review',
        kind: 'risk',
        severity: 'high',
        updated_at: '2026-03-11T05:00:00Z',
      }),
      item({
        text: 'future aiops lane',
        source: 'founder_idea',
        kind: 'proposal',
        severity: 'low',
        updated_at: '2026-03-11T05:30:00Z',
      }),
      item({
        text: 'misc unresolved note',
        source: 'manual',
        kind: 'agenda',
        severity: 'medium',
        updated_at: '2026-03-10T02:00:00Z',
      }),
    ],
    deferred: [{}],
    accepted: [{}, {}],
  }, { now: NOW, staleWindowHours: STALE_REVIEW_WINDOW_HOURS });

  assert.equal(triage.metrics.open, 5);
  assert.equal(triage.metrics.blockers, 2);
  assert.equal(triage.metrics.residue, 1);
  assert.equal(triage.metrics.strategic, 1);
  assert.equal(triage.metrics.other, 1);
  assert.equal(triage.metrics.deferred, 1);
  assert.equal(triage.metrics.accepted, 2);
  assert.equal(triage.activeBlockers[0].text, 'fresh failed agent job');
  assert.equal(triage.runtimeResidue[0].text, 'old failed verifier');
  assert.equal(triage.strategicBacklog[0].text, 'future aiops lane');
  assert.equal(triage.otherUnresolved[0].text, 'misc unresolved note');
});
