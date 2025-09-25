import { type NextRequest, NextResponse } from "next/server";

// This would typically update your database
// For now, we'll simulate the reordering logic
export async function POST(request: NextRequest) {
  try {
    const environments = await request.json();

    const response = await fetch(
      `${
        process.env.NEXT_PUBLIC_BASE_URL || "http://localhost:8080"
      }/environments/reorder`,
      {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(environments),
      }
    );

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`Failed to reorder environments: ${errorText}`);
    }

    return new NextResponse(null, { status: 204 });
  } catch (error) {
    return NextResponse.json(
      { error: "Failed to update environment order" },
      { status: 500 }
    );
  }
}
