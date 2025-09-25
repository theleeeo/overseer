import { NextResponse } from "next/server";

export async function GET() {
  try {
    const res = await fetch(
      `${
        process.env.NEXT_PUBLIC_BASE_URL || "http://localhost:8080"
      }/environments`,
      { cache: "no-store" }
    );
    if (!res.ok) {
      throw new Error("Failed to fetch environments from backend");
    }
    const data = await res.json();
    return NextResponse.json({ environments: data });
  } catch (error) {
    return NextResponse.json(
      { error: "Failed to fetch environments" },
      { status: 500 }
    );
  }
}
