import { ArrowLeft, ArrowRight, Check, PackageSearch } from "lucide-react";
import { useState, type FormEvent } from "react";
import { Link, Navigate, useNavigate } from "react-router-dom";
import { useAuth } from "../state/auth-context";
const categories = [
  "CRT",
  "LCD Panel",
  "PCB",
  "Cable",
  "Battery",
  "Motor",
  "Magnet-bearing Assembly",
  "Mixed Plastics from EEE",
  "Other E-Waste",
];
export default function RegisterPage() {
  const { user, register } = useAuth();
  const navigate = useNavigate();
  const [step, setStep] = useState(1);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [form, setForm] = useState({
    name: "",
    phone: "",
    email: "",
    password: "",
    confirm: "",
    companyName: "",
    facilityName: "",
    businessType: "E-waste Recycler / Processing Unit",
    gstNumber: "",
    registrationNumber: "",
    authorizationAuthority: "",
    street: "",
    city: "",
    state: "Haryana",
    pincode: "",
    acceptedMaterials: ["PCB", "Cable"],
    processingCapacityKg: "",
    pickupAvailable: true,
    serviceArea: "Kurukshetra, Karnal",
    termsAccepted: false,
  });
  if (user) return <Navigate to="/dashboard" replace />;
  const update = (k: string, v: string | boolean | string[]) =>
    setForm((f) => ({ ...f, [k]: v }));
  const next = () => {
    setError("");
    if (
      step === 1 &&
      (!form.name ||
        !form.phone ||
        !form.email ||
        form.password.length < 10 ||
        form.password !== form.confirm)
    ) {
      setError(
        "Complete all contact fields. Passwords must match and be at least 10 characters.",
      );
      return;
    }
    if (
      step === 2 &&
      (!form.companyName ||
        !form.facilityName ||
        !form.businessType ||
        !form.street ||
        !form.city ||
        !form.state ||
        !form.pincode)
    ) {
      setError("Complete the facility and address fields.");
      return;
    }
    setStep((s) => Math.min(3, s + 1));
  };
  const submit = async (e: FormEvent) => {
    e.preventDefault();
    if (step < 3) {
      next();
      return;
    }
    if (!form.termsAccepted) {
      setError("Please accept the terms and privacy agreement.");
      return;
    }
    setBusy(true);
    setError("");
    try {
      await register({
        name: form.name,
        phone: form.phone,
        email: form.email,
        password: form.password,
        companyName: form.companyName,
        facilityName: form.facilityName,
        businessType: form.businessType,
        gstNumber: form.gstNumber,
        registrationNumber: form.registrationNumber,
        authorizationAuthority: form.authorizationAuthority,
        address: {
          street: form.street,
          city: form.city,
          state: form.state,
          pincode: form.pincode,
        },
        acceptedMaterials: form.acceptedMaterials,
        processingCapacityKg: form.processingCapacityKg
          ? Number(form.processingCapacityKg)
          : null,
        pickupAvailable: form.pickupAvailable,
        serviceArea: form.serviceArea
          .split(",")
          .map((x) => x.trim())
          .filter(Boolean),
        termsAccepted: true,
      });
      navigate("/dashboard", { replace: true });
    } catch (e) {
      setError(e instanceof Error ? e.message : "Registration failed");
    } finally {
      setBusy(false);
    }
  };
  return (
    <div className="register-page">
      <div className="register-shell">
        <div className="register-brand">
          <PackageSearch />
          <strong>KabadiConnect</strong>
          <Link to="/login">Already registered? Sign in</Link>
        </div>
        <div className="stepper">
          {["Contact", "Facility", "Authorization & Materials"].map(
            (name, i) => (
              <div
                className={`step ${step >= i + 1 ? "active" : ""}`}
                key={name}
              >
                <span>{step > i + 1 ? <Check size={16} /> : i + 1}</span>
                <div>
                  <strong>{name}</strong>
                  <small>
                    {i === 0
                      ? "Account"
                      : i === 1
                        ? "Recycler facility"
                        : "Development verification"}
                  </small>
                </div>
              </div>
            ),
          )}
        </div>
        <form className="register-card" onSubmit={submit}>
          <span className="eyebrow">Step {step} of 3</span>
          <h1>
            {step === 1
              ? "Create your recycler account"
              : step === 2
                ? "Recycler facility details"
                : "Authorization and e-waste scope"}
          </h1>
          <p>
            {step === 3
              ? "In local development, new accounts are marked DEMO VERIFIED for testing only. Production accounts remain pending verification."
              : "Use business details for the recycler-side workspace."}
          </p>
          {error && (
            <div className="form-error" role="alert">
              {error}
            </div>
          )}
          {step === 1 && (
            <div className="form-grid">
              <label className="span-2">
                Contact person name
                <input
                  value={form.name}
                  onChange={(e) => update("name", e.target.value)}
                  required
                />
              </label>
              <label>
                Phone
                <input
                  value={form.phone}
                  onChange={(e) => update("phone", e.target.value)}
                  required
                />
              </label>
              <label>
                Email
                <input
                  type="email"
                  value={form.email}
                  onChange={(e) => update("email", e.target.value)}
                  required
                />
              </label>
              <label>
                Password
                <input
                  type="password"
                  value={form.password}
                  onChange={(e) => update("password", e.target.value)}
                  placeholder="Minimum 10 characters"
                  required
                />
              </label>
              <label>
                Confirm password
                <input
                  type="password"
                  value={form.confirm}
                  onChange={(e) => update("confirm", e.target.value)}
                  required
                />
              </label>
            </div>
          )}
          {step === 2 && (
            <div className="form-grid">
              <label>
                Company name
                <input
                  value={form.companyName}
                  onChange={(e) => update("companyName", e.target.value)}
                  required
                />
              </label>
              <label>
                Facility name
                <input
                  value={form.facilityName}
                  onChange={(e) => update("facilityName", e.target.value)}
                  required
                />
              </label>
              <label>
                Business type
                <select
                  value={form.businessType}
                  onChange={(e) => update("businessType", e.target.value)}
                >
                  <option>E-waste Recycler / Processing Unit</option>
                  <option>Authorized Dismantler</option>
                  <option>Refurbisher</option>
                  <option>Material Recovery Facility</option>
                  <option>Aggregator / Bulk Buyer</option>
                </select>
              </label>
              <label>
                GST number <small>optional</small>
                <input
                  value={form.gstNumber}
                  onChange={(e) => update("gstNumber", e.target.value)}
                />
              </label>
              <label className="span-2">
                Street address
                <input
                  value={form.street}
                  onChange={(e) => update("street", e.target.value)}
                  required
                />
              </label>
              <label>
                City
                <input
                  value={form.city}
                  onChange={(e) => update("city", e.target.value)}
                  required
                />
              </label>
              <label>
                State
                <input
                  value={form.state}
                  onChange={(e) => update("state", e.target.value)}
                  required
                />
              </label>
              <label>
                PIN code
                <input
                  value={form.pincode}
                  onChange={(e) => update("pincode", e.target.value)}
                  required
                />
              </label>
            </div>
          )}
          {step === 3 && (
            <>
              <div className="form-grid">
                <label>
                  Authorization / registration reference{" "}
                  <small>optional in dev</small>
                  <input
                    value={form.registrationNumber}
                    onChange={(e) =>
                      update("registrationNumber", e.target.value)
                    }
                    placeholder="Do not enter fake production credentials"
                  />
                </label>
                <label>
                  Issuing authority <small>optional</small>
                  <input
                    value={form.authorizationAuthority}
                    onChange={(e) =>
                      update("authorizationAuthority", e.target.value)
                    }
                  />
                </label>
              </div>
              <fieldset className="material-picker">
                <legend>E-waste material categories accepted</legend>
                <div>
                  {categories.map((m) => (
                    <label
                      className={`material-chip ${form.acceptedMaterials.includes(m) ? "selected" : ""}`}
                      key={m}
                    >
                      <input
                        type="checkbox"
                        checked={form.acceptedMaterials.includes(m)}
                        onChange={() =>
                          update(
                            "acceptedMaterials",
                            form.acceptedMaterials.includes(m)
                              ? form.acceptedMaterials.filter((x) => x !== m)
                              : [...form.acceptedMaterials, m],
                          )
                        }
                      />
                      {form.acceptedMaterials.includes(m) && (
                        <Check size={15} />
                      )}{" "}
                      {m}
                    </label>
                  ))}
                </div>
              </fieldset>
              <div className="form-grid">
                <label>
                  Approx. processing capacity <small>kg/month, optional</small>
                  <input
                    type="number"
                    min="0"
                    value={form.processingCapacityKg}
                    onChange={(e) =>
                      update("processingCapacityKg", e.target.value)
                    }
                  />
                </label>
                <label>
                  Pickup service area
                  <input
                    value={form.serviceArea}
                    onChange={(e) => update("serviceArea", e.target.value)}
                    placeholder="Comma-separated cities"
                  />
                </label>
              </div>
              <label className="terms-check">
                <input
                  type="checkbox"
                  checked={form.pickupAvailable}
                  onChange={(e) => update("pickupAvailable", e.target.checked)}
                />
                <span>Recycler pickup is available</span>
              </label>
              <label className="terms-check">
                <input
                  type="checkbox"
                  checked={form.termsAccepted}
                  onChange={(e) => update("termsAccepted", e.target.checked)}
                />
                <span>I agree to the Terms of Service and Privacy Policy.</span>
              </label>
            </>
          )}
          <div className="register-actions">
            {step > 1 ? (
              <button
                type="button"
                className="secondary-button"
                onClick={() => setStep((s) => s - 1)}
              >
                <ArrowLeft size={17} />
                Back
              </button>
            ) : (
              <Link className="secondary-button" to="/login">
                <ArrowLeft size={17} />
                Back to login
              </Link>
            )}
            <button className="primary-button" disabled={busy}>
              {step < 3 ? (
                <>
                  Continue <ArrowRight size={17} />
                </>
              ) : busy ? (
                "Creating account..."
              ) : (
                <>
                  Create recycler account <ArrowRight size={17} />
                </>
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
