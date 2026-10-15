import { useState } from 'react';
import { proposalsApi } from '../../../api/proposals';
import type { components } from '../../../api/schema';
import { useReadApi } from '../../../hooks/useReadApi';
import { QueryState } from '../../../components/QueryState';
import { GoalRevision } from './GoalRevision';

export function FixedGoalReference({ goal }: { goal: components['schemas']['S2GoalReferenceV1'] }) {
  const [opened, setOpened] = useState(false);
  const available = !!goal.id && goal.revision > 0;
  const fixed = useReadApi(signal => proposalsApi.goal(goal.id, goal.revision, signal), { enabled: opened && available, key: `${goal.id}@${goal.revision}` });
  if (!available) return <p>当前提案尚未选择固定目标版本。</p>;
  return <details onToggle={event => setOpened(event.currentTarget.open)}><summary>展开提案引用的固定目标 r{goal.revision}</summary>
    {opened && <section aria-label="提案引用的固定目标"><QueryState state={fixed}>{({ goal: snapshot }) => <GoalRevision goal={snapshot} />}</QueryState></section>}
  </details>;
}
