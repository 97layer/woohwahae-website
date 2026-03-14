import { NextResponse } from 'next/server';

import { getNeuralGraphView } from '../../../../lib/runtime/view-model';

export async function GET() {
  try {
    const graph = await getNeuralGraphView();
    return NextResponse.json({ graph });
  } catch (error) {
    return NextResponse.json(
      { error: error instanceof Error ? error.message : 'runtime unavailable' },
      { status: 502 },
    );
  }
}
