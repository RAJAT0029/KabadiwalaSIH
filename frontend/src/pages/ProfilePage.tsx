import { BadgeCheck, Check, Save, ShieldAlert } from "lucide-react";
import { useMemo, useState } from "react";
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
export default function ProfilePage() {
  const { user, updateProfile } = useAuth();
  const [busy, setBusy] = useState(false);
  const [success, setSuccess] = useState("");
  const [error, setError] = useState("");
  const initial = useMemo(
    () => ({
      name: user?.name || "",
      companyName: user?.companyName || "",
      facilityName: user?.facilityName || "",
      businessType: user?.businessType || "",
      gstNumber: user?.gstNumber || "",
      street: user?.address.street || "",
      city: user?.address.city || "",
      state: user?.address.state || "",
      pincode: user?.address.pincode || "",
      acceptedMaterials: user?.acceptedMaterials || [],
      processingCapacityKg: user?.processingCapacityKg
        ? String(user.processingCapacityKg)
        : "",
      registrationNumber: user?.authorization?.registrationNumber || "",
      authority: user?.authorization?.authority || "",
      documentReference: user?.authorization?.documentReference || "",
      pickupAvailable: user?.pickup?.available || false,
      serviceArea: (user?.pickup?.serviceArea || []).join(", "),
    }),
    [user],
  );
  const [form, setForm] = useState(initial);
  if (!user) return null;
  const update = (k: string, v: string | boolean | string[]) =>
    setForm((f) => ({ ...f, [k]: v }));
  const save = async () => {
    setBusy(true);
    setError("");
    setSuccess("");
    try {
      await updateProfile({
        name: form.name,
        companyName: form.companyName,
        facilityName: form.facilityName,
        businessType: form.businessType,
        gstNumber: form.gstNumber,
        address: {
          street: form.street,
          city: form.city,
          state: form.state,
          pincode: form.pincode,
        },
        acceptedMaterials: form.acceptedMaterials,
        processingCapacityKg: form.processingCapacityKg
          ? Number(form.processingCapacityKg)
          : undefined,
        authorization: {
          registrationNumber: form.registrationNumber,
          authority: form.authority,
          documentReference: form.documentReference,
        },
        pickup: {
          available: form.pickupAvailable,
          serviceArea: form.serviceArea
            .split(",")
            .map((x) => x.trim())
            .filter(Boolean),
        },
      });
      setSuccess(
        "Recycler profile updated. Authorization status itself cannot be self-promoted from this form.",
      );
    } catch (e) {
      setError(e instanceof Error ? e.message : "Profile update failed");
    } finally {
      setBusy(false);
    }
  };
  const auth = user.effectiveAuthorizationStatus || "pending_verification";
  return (
    <div>
      <div className="page-heading">
        <div>
          <span className="eyebrow">Recycler authorization dataset</span>
          <h1>Recycler Profile</h1>
          <p>
            Facility, accepted materials, pickup capability and authorization
            references used by recycler matching and offer eligibility.
          </p>
        </div>
        <button className="primary-button" onClick={save} disabled={busy}>
          <Save size={17} />
          {busy ? "Saving…" : "Save changes"}
        </button>
      </div>
      <div
        className={`authorization-banner ${user.participationAllowed ? "ok" : "warn"}`}
      >
        {user.participationAllowed ? <BadgeCheck /> : <ShieldAlert />}
        <div>
          <strong>
            {user.authorization?.isDemo
              ? `DEVELOPMENT AUTHORIZATION — ${auth.replaceAll("_", " ").toUpperCase()}`
              : auth.replaceAll("_", " ").toUpperCase()}
          </strong>
          <span>
            {user.authorization?.isDemo
              ? "Not a real government authorization. Production verification must use a validated source/process."
              : "Authorization status controls recycler participation."}
          </span>
        </div>
      </div>
      {success && (
        <div className="success-banner">
          <Check />
          {success}
        </div>
      )}
      {error && <div className="inline-error">{error}</div>}
      <div className="profile-grid">
        <section className="profile-section">
          <h2>Facility & contact</h2>
          <p>
            {user.email} · {user.phone}
          </p>
          <div className="form-grid">
            <label>
              Contact person
              <input
                value={form.name}
                onChange={(e) => update("name", e.target.value)}
              />
            </label>
            <label>
              Company name
              <input
                value={form.companyName}
                onChange={(e) => update("companyName", e.target.value)}
              />
            </label>
            <label>
              Facility name
              <input
                value={form.facilityName}
                onChange={(e) => update("facilityName", e.target.value)}
              />
            </label>
            <label>
              Business type
              <input
                value={form.businessType}
                onChange={(e) => update("businessType", e.target.value)}
              />
            </label>
            <label>
              GST number
              <input
                value={form.gstNumber}
                onChange={(e) => update("gstNumber", e.target.value)}
              />
            </label>
            <label>
              Processing capacity kg/month
              <input
                type="number"
                min="0"
                value={form.processingCapacityKg}
                onChange={(e) => update("processingCapacityKg", e.target.value)}
              />
            </label>
          </div>
        </section>
        <section className="profile-section">
          <h2>Facility address</h2>
          <div className="form-grid">
            <label className="span-2">
              Street
              <input
                value={form.street}
                onChange={(e) => update("street", e.target.value)}
              />
            </label>
            <label>
              City
              <input
                value={form.city}
                onChange={(e) => update("city", e.target.value)}
              />
            </label>
            <label>
              State
              <input
                value={form.state}
                onChange={(e) => update("state", e.target.value)}
              />
            </label>
            <label>
              PIN
              <input
                value={form.pincode}
                onChange={(e) => update("pincode", e.target.value)}
              />
            </label>
          </div>
        </section>
        <section className="profile-section">
          <h2>Authorization references</h2>
          <p>
            Valid from:{" "}
            {user.authorization.validFrom
              ? new Date(user.authorization.validFrom).toLocaleDateString()
              : "Not recorded"}{" "}
            · Valid until:{" "}
            {user.authorization.validUntil
              ? new Date(user.authorization.validUntil).toLocaleDateString()
              : "Not recorded"}
          </p>
          <p className="section-note">
            Changing verified references requires verification again. Only the
            verification service can change validity dates.
          </p>
          <div className="form-grid">
            <label>
              Registration number
              <input
                value={form.registrationNumber}
                onChange={(e) => update("registrationNumber", e.target.value)}
              />
            </label>
            <label>
              Issuing authority
              <input
                value={form.authority}
                onChange={(e) => update("authority", e.target.value)}
              />
            </label>
            <label className="span-2">
              Document reference
              <input
                value={form.documentReference}
                onChange={(e) => update("documentReference", e.target.value)}
                placeholder="Secure object/document reference when available"
              />
            </label>
          </div>
        </section>
        <section className="profile-section">
          <h2>Accepted e-waste materials</h2>
          <div className="material-picker">
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
                  {form.acceptedMaterials.includes(m) && <Check size={15} />}{" "}
                  {m}
                </label>
              ))}
            </div>
          </div>
        </section>
        <section className="profile-section">
          <h2>Pickup capability</h2>
          <label className="terms-check">
            <input
              type="checkbox"
              checked={form.pickupAvailable}
              onChange={(e) => update("pickupAvailable", e.target.checked)}
            />
            <span>Facility can arrange pickup</span>
          </label>
          <label>
            Service area
            <input
              value={form.serviceArea}
              onChange={(e) => update("serviceArea", e.target.value)}
              placeholder="Comma-separated cities/areas"
            />
          </label>
        </section>
      </div>
    </div>
  );
}
