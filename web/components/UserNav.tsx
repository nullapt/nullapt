import { getCurrentUser } from "@/lib/auth";

export async function UserNav() {
  const user = await getCurrentUser();

  if (!user) {
    return (
      <a
        href="/login"
        style={{ background: "var(--accent)", color: "#000" }}
        className="rounded px-3 py-1 text-xs font-semibold hover:opacity-90 transition-opacity"
      >
        sign in
      </a>
    );
  }

  return (
    <div className="flex items-center gap-3 pl-4" style={{ borderLeft: "1px solid var(--border)" }}>
      <a href="/settings/tokens" className="flex items-center gap-2 hover:text-white transition-colors">
        {user.avatar_url && (
          // eslint-disable-next-line @next/next/no-img-element
          <img
            src={user.avatar_url}
            alt={user.username}
            width={20}
            height={20}
            style={{ borderRadius: "50%" }}
          />
        )}
        <span style={{ color: "var(--accent)" }}>@{user.username}</span>
      </a>
      <form action="/auth/logout" method="POST">
        <button
          type="submit"
          className="text-xs hover:text-white transition-colors"
          style={{ color: "var(--muted)" }}
        >
          logout
        </button>
      </form>
    </div>
  );
}
