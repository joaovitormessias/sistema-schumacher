import { afterEach, describe, expect, it, vi } from "vitest";

const createClient = vi.fn();

vi.mock("@supabase/supabase-js", () => ({
  createClient,
}));

describe("supabase service", () => {
  afterEach(() => {
    vi.unstubAllEnvs();
    vi.resetModules();
    vi.clearAllMocks();
  });

  it("treats invalid VITE_SUPABASE_URL as not configured", async () => {
    vi.stubEnv("VITE_SUPABASE_URL", "not-a-url");
    vi.stubEnv("VITE_SUPABASE_ANON_KEY", "anon-key");

    const mod = await import("../supabase");

    expect(mod.isSupabaseConfigured()).toBe(false);
    expect(mod.getSupabaseClient()).toBeNull();
    expect(createClient).not.toHaveBeenCalled();
  });

  it("creates the client when URL and key are valid", async () => {
    vi.stubEnv("VITE_SUPABASE_URL", "https://example.supabase.co");
    vi.stubEnv("VITE_SUPABASE_ANON_KEY", "anon-key");
    createClient.mockReturnValue({ id: "client" });

    const mod = await import("../supabase");

    expect(mod.isSupabaseConfigured()).toBe(true);
    expect(mod.getSupabaseClient()).toEqual({ id: "client" });
    expect(createClient).toHaveBeenCalledWith("https://example.supabase.co", "anon-key");
  });
});
