import { NextResponse } from "next/server";

export async function GET() {
  try {
    const res = await fetch(
      `${process.env.NEXT_PUBLIC_BASE_URL || "http://localhost:8080"}/versions`,
      { cache: "no-store" }
    );
    if (!res.ok) {
      throw new Error("Failed to fetch version data");
    }
    const data = await res.json();

    return NextResponse.json({ cells: data });
  } catch (error) {
    return NextResponse.json(
      { error: "Failed to fetch version data" },
      { status: 500 }
    );
  }
}
