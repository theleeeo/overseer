import { type NextRequest, NextResponse } from "next/server";

// This would typically update your database
// For now, we'll simulate the reordering logic
export async function POST(request: NextRequest) {
  try {
    const applications = await request.json();

    const response = await fetch(
      `${
        process.env.NEXT_PUBLIC_BASE_URL || "http://localhost:8080"
      }/applications`,
      {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(applications),
      }
    );

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`Failed to reorder applications: ${errorText}`);
    }

    return new NextResponse(null, { status: 204 });
  } catch (error) {
    console.error("[v0] Failed to update application order:", error);
    return NextResponse.json(
      { error: "Failed to update application order" },
      { status: 500 }
    );
  }
}
