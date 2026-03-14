import NeuralGraphDemo from '../../../components/neural-graph-demo';
import { getNeuralGraphView } from '../../../lib/runtime/view-model';

export const metadata = {
  title: 'Layer OS Neural Map',
  description: 'Live runtime graph demo for Layer OS.',
};

export default async function NeuralGraphPage() {
  const initialGraph = await getNeuralGraphView();
  return <NeuralGraphDemo initialGraph={initialGraph} />;
}
