import { ArrowRight, CheckCircle2, PackageSearch } from "lucide-react";
import { useState, type FormEvent } from "react";
import { Link, Navigate, useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "../state/auth-context";
export default function LoginPage() {
  const { user, login } = useAuth();
  const nav = useNavigate();
  const loc = useLocation();
  const [identifier, setIdentifier] = useState("");
  const [password, setPassword] = useState("");
  const [remember, setRemember] = useState(true);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  if (user) return <Navigate to="/dashboard" replace />;
  const submit = async (e: FormEvent) => {
    e.preventDefault();
    setBusy(true);
    setError("");
    try {
      await login(identifier, password, remember);
      nav((loc.state as { from?: string } | null)?.from || "/dashboard", {
        replace: true,
      });
    } catch (e) {
      setError(e instanceof Error ? e.message : "Login failed");
    } finally {
      setBusy(false);
    }
  };
  return (
    <div className="auth-page">
      <section className="auth-showcase">
        <div className="auth-brand">
          <PackageSearch />
          KabadiConnect
        </div>
        <div className="auth-showcase-copy">
          <span className="eyebrow light">
            E-waste formalization for recyclers
          </span>
          <h1>
            Source e-waste through transparent, traceable recycler workflows.
          </h1>
          <p>
            Discover collector-created e-waste lots, compare indicative price
            records, send offers and retain traceable handover history.
          </p>
          <div className="benefit-list">
            <span>
              <CheckCircle2 />
              Authorization-aware recycler access
            </span>
            <span>
              <CheckCircle2 />
              Clear offer workflow
            </span>
            <span>
              <CheckCircle2 />
              Traceable e-waste handovers
            </span>
          </div>
        </div>
        <small>
          Built for formal e-waste transfer, not generic scrap trading.
        </small>
      </section>
      <section className="auth-panel">
        <form className="auth-card" onSubmit={submit}>
          <div className="mobile-auth-logo">
            <PackageSearch />
            KabadiConnect
          </div>
          <span className="eyebrow">Welcome back</span>
          <h2>Sign in to your recycler account</h2>
          <p>Continue to your e-waste recycler workspace.</p>
          {error && (
            <div className="form-error" role="alert">
              {error}
            </div>
          )}
          <label>
            Email or phone
            <input
              value={identifier}
              onChange={(e) => setIdentifier(e.target.value)}
              placeholder="you@company.com"
              autoComplete="username"
              required
            />
          </label>
          <label>
            Password
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Enter your password"
              autoComplete="current-password"
              required
            />
          </label>
          <div className="form-row">
            <label className="checkbox">
              <input
                type="checkbox"
                checked={remember}
                onChange={(e) => setRemember(e.target.checked)}
              />
              Remember me
            </label>
            <span title="Password recovery is not yet implemented">
              Recovery coming soon
            </span>
          </div>
          <button className="primary-button full" disabled={busy}>
            {busy ? (
              "Signing in..."
            ) : (
              <>
                Sign in <ArrowRight size={17} />
              </>
            )}
          </button>
          <p className="auth-switch">
            New to KabadiConnect?{" "}
            <Link to="/register">Create Recycler Account</Link>
          </p>
        </form>
      </section>
    </div>
  );
}
