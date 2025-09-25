import { NextResponse } from "next/server";

export async function POST(request: Request) {
  try {
    const body = await request.json();

    const res = await fetch(
      `${
        process.env.NEXT_PUBLIC_BASE_URL || "http://localhost:8080"
      }/environments`,
      {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      }
    );
    if (!res.ok) {
      throw new Error("Failed to create environment in backend");
    }

    return NextResponse.json(await res.json());
  } catch (error) {
    return NextResponse.json(
      { error: "Failed to create environment" },
      { status: 500 }
    );
  }
}

export async function PUT(request: Request) {
  try {
    const body = await request.json();

    const res = await fetch(
      `${
        process.env.NEXT_PUBLIC_BASE_URL || "http://localhost:8080"
      }/environments/${body.id}`,
      {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      }
    );
    if (!res.ok) {
      throw new Error("Failed to update environment in backend");
    }

    return NextResponse.json(await res.json());
  } catch (error) {
    return NextResponse.json(
      { error: "Failed to update environment" },
      { status: 500 }
    );
  }
}

export async function DELETE(request: Request) {
  try {
    const { searchParams } = new URL(request.url);
    const id = searchParams.get("id");
    if (!id) {
      return NextResponse.json(
        { error: "Environment ID is required" },
        { status: 400 }
      );
    }

    const res = await fetch(
      `${
        process.env.NEXT_PUBLIC_BASE_URL || "http://localhost:8080"
      }/environments/${id}`,
      {
        method: "DELETE",
      }
    );
    if (!res.ok) {
      throw new Error("Failed to delete environment in backend");
    }

    return NextResponse.json({ message: "Environment deleted" });
  } catch (error) {
    return NextResponse.json(
      { error: "Failed to delete environment" },
      { status: 500 }
    );
  }
}
