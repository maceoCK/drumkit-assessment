import { useEffect, useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useForm, FormProvider } from 'react-hook-form'
import { z } from 'zod'
import { zodResolver } from '@hookform/resolvers/zod'

type CreateLoadModalProps = {
  open: boolean
  onClose: () => void
  onSuccess: () => Promise<void> | void
}

const optStr = z.preprocess((v) => (typeof v === 'string' && v.trim() === '' ? undefined : v), z.string().optional())
const optEmail = z.preprocess((v) => (typeof v === 'string' && v.trim() === '' ? undefined : v), z.string().email('Invalid email').optional())

const partySchema = z.object({
  turvoId: z.number().int().optional(),
  externalTMSId: optStr,
  name: z.string().min(1, 'Required'),
  addressLine1: z.string().min(1, 'Required'),
  addressLine2: optStr,
  city: z.string().min(1, 'Required'),
  state: z.string().min(1, 'Required'),
  zipcode: z.string().min(1, 'Required'),
  country: z.string().min(1, 'Required').default('US'),
  contact: optStr,
  phone: optStr,
  email: optEmail,
  refNumber: optStr,
})

const stopSchema = partySchema.extend({
  businessHours: optStr,
  timezone: optStr,
  warehouseId: optStr,
})

const carrierSchema = z.object({
  mcNumber: optStr,
  dotNumber: optStr,
  name: optStr,
  phone: optStr,
  dispatcher: optStr,
  sealNumber: optStr,
  scac: optStr,
  firstDriverName: optStr,
  firstDriverPhone: optStr,
  secondDriverName: optStr,
  secondDriverPhone: optStr,
  email: optStr,
  dispatchCity: optStr,
  dispatchState: optStr,
  externalTMSTruckId: optStr,
  externalTMSTrailerId: optStr,
  confirmationSentTime: optStr,
  confirmationReceivedTime: optStr,
  dispatchedTime: optStr,
  expectedPickupTime: optStr,
  pickupStart: optStr,
  pickupEnd: optStr,
  expectedDeliveryTime: optStr,
  deliveryStart: optStr,
  deliveryEnd: optStr,
  signedBy: optStr,
  externalTMSId: optStr,
})

const rateDataSchema = z.object({
  customerRateType: optStr,
  customerNumHours: optStr,
  customerLhRateUsd: optStr,
  fscPercent: optStr,
  fscPerMile: optStr,
  carrierRateType: optStr,
  carrierNumHours: optStr,
  carrierLhRateUsd: optStr,
  carrierMaxRate: optStr,
  netProfitUsd: optStr,
  profitPercent: optStr,
})

const specificationsSchema = z.object({
  minTempFahrenheit: optStr,
  maxTempFahrenheit: optStr,
  liftgatePickup: z.boolean().optional(),
  liftgateDelivery: z.boolean().optional(),
  insidePickup: z.boolean().optional(),
  insideDelivery: z.boolean().optional(),
  tarps: z.boolean().optional(),
  oversized: z.boolean().optional(),
  hazmat: z.boolean().optional(),
  straps: z.boolean().optional(),
  permits: z.boolean().optional(),
  escorts: z.boolean().optional(),
  seal: z.boolean().optional(),
  customBonded: z.boolean().optional(),
  labor: z.boolean().optional(),
  inPalletCount: z.string().optional(),
  outPalletCount: z.string().optional(),
  numCommodities: z.string().optional(),
  totalWeight: z.string().optional(),
  billableWeight: z.string().optional(),
  poNums: z.string().optional(),
  operator: z.string().optional(),
  routeMiles: z.string().optional(),
})

const formSchema = z.object({
  externalTMSLoadID: z.string().min(1, 'Required'),
  status: z.string().min(1, 'Required').default('NEW'),
  freightLoadID: z.string().optional(),
  customer: partySchema,
  billToEnabled: z.boolean().optional(),
  billTo: partySchema.optional(),
  pickup: stopSchema,
  consignee: stopSchema,
  scheduling: z.object({
    readyTime: z.string().min(1, 'Pickup time required'),
    mustDeliver: z.string().min(1, 'Delivery time required'),
    pickupApptTime: z.string().optional(),
    pickupApptNote: z.string().optional(),
    consigneeApptTime: z.string().optional(),
    consigneeApptNote: z.string().optional(),
    timezone: z.string().optional(),
  }),
  carrierEnabled: z.boolean().optional(),
  carrier: carrierSchema.optional(),
  rateDataEnabled: z.boolean().optional(),
  rateData: rateDataSchema.optional(),
  specsEnabled: z.boolean().optional(),
  specifications: specificationsSchema.optional(),
})

type FormValues = z.infer<typeof formSchema>

export default function CreateLoadModal({ open, onClose, onSuccess }: CreateLoadModalProps) {
  const [step, setStep] = useState(0)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const methods = useForm<FormValues>({
    resolver: zodResolver(formSchema) as any,
    defaultValues: {
      externalTMSLoadID: '',
      status: 'NEW',
      customer: { name: '', addressLine1: '', city: '', state: '', zipcode: '', country: 'US' },
      pickup: { name: '', addressLine1: '', city: '', state: '', zipcode: '', country: 'US' },
      consignee: { name: '', addressLine1: '', city: '', state: '', zipcode: '', country: 'US' },
      scheduling: { readyTime: '', mustDeliver: '' },
      billToEnabled: false,
      carrierEnabled: false,
      rateDataEnabled: false,
      specsEnabled: false,
    },
    mode: 'onBlur',
  })

  const stepFieldPaths: string[][] = [
    ['externalTMSLoadID', 'status'],
    ['customer.name', 'customer.addressLine1', 'customer.city', 'customer.state', 'customer.zipcode'],
    ['pickup.name', 'pickup.addressLine1', 'pickup.city', 'pickup.state', 'pickup.zipcode'],
    ['consignee.name', 'consignee.addressLine1', 'consignee.city', 'consignee.state', 'consignee.zipcode'],
    ['scheduling.readyTime', 'scheduling.mustDeliver'],
  ]

  // Optional groups state
  const billToEnabled = methods.watch('billToEnabled')
  const carrierEnabled = methods.watch('carrierEnabled')
  const rateDataEnabled = methods.watch('rateDataEnabled')
  const specsEnabled = methods.watch('specsEnabled')

  // Customers dropdown state
  const [customers, setCustomers] = useState<Array<{ id: number; name: string }>>([])
  const [loadingCustomers, setLoadingCustomers] = useState(false)
  useEffect(() => {
    if (!open) return
    ;(async () => {
      try {
        setLoadingCustomers(true)
        const r = await fetch('/api/customers?pageSize=50')
        if (!r.ok) throw new Error('Failed customers')
        const data = await r.json()
        const items = Array.isArray(data) ? data : (data?.items ?? [])
        setCustomers(items)
      } catch (e) {
        console.warn('[CreateLoadModal] customers fetch failed', e)
      } finally {
        setLoadingCustomers(false)
      }
    })()
  }, [open])

  // Unregister optional groups when disabled to avoid validation on empty objects
  useEffect(() => {
    if (!billToEnabled) {
      methods.unregister('billTo')
    }
  }, [billToEnabled])
  useEffect(() => {
    if (!carrierEnabled) {
      methods.unregister('carrier')
    }
  }, [carrierEnabled])
  useEffect(() => {
    if (!rateDataEnabled) {
      methods.unregister('rateData')
    }
  }, [rateDataEnabled])
  useEffect(() => {
    if (!specsEnabled) {
      methods.unregister('specifications')
    }
  }, [specsEnabled])


  function SectionTitle({ children }: { children: string }) {
    return <h3 className="sm:col-span-2 mt-4 text-sm font-semibold text-gray-700">{children}</h3>
  }

  function Field({ name, label, type = 'text', required = false }: { name: any; label: string; type?: string; required?: boolean }) {
    const { register, formState: { errors } } = methods
    const err = String(name).split('.').reduce((acc: any, key: string) => (acc ? acc[key] : undefined), errors as any)
    return (
      <div className="grid gap-1">
        <Label htmlFor={name}>{label}</Label>
        <Input id={name} type={type} {...register(name, { required })} />
        {err?.message && <span className="text-xs text-red-600">{String(err.message)}</span>}
      </div>
    )
  }

  function Checkbox({ name, label }: { name: any; label: string }) {
    const { register } = methods
    return (
      <label className="flex items-center gap-2 text-sm">
        <input type="checkbox" {...register(name)} />
        <span>{label}</span>
      </label>
    )
  }

  async function nextStep() {
    const fields = stepFieldPaths[step] || []
    console.log('[CreateLoadModal] nextStep -> validating fields:', fields)
    const ok = await methods.trigger(fields as any)
    console.log('[CreateLoadModal] nextStep -> valid:', ok)
    if (!ok) return
    setStep((s) => Math.min(s + 1, 5))
  }

  function prevStep() {
    console.log('[CreateLoadModal] prevStep')
    setStep((s) => Math.max(s - 1, 0))
  }

  function toISO(value?: string) {
    const v = (value || '').trim()
    if (!v) return undefined
    try { return new Date(v).toISOString() } catch { return undefined }
  }

  function toNum(value?: string): number | undefined {
    const v = (value || '').trim()
    if (!v) return undefined
    const n = Number(v)
    return Number.isFinite(n) ? n : undefined
  }

  async function onSubmit(values: FormValues) {
    setSubmitting(true)
    setError(null)
    try {
      console.log('[CreateLoadModal] submit -> values:', values)
      const payload: any = {
        externalTMSLoadID: values.externalTMSLoadID,
        status: values.status,
        customer: values.customer,
        pickup: { ...values.pickup, readyTime: toISO(values.scheduling.readyTime), apptTime: toISO(values.scheduling.pickupApptTime), apptNote: values.scheduling.pickupApptNote, timezone: values.scheduling.timezone },
        consignee: { ...values.consignee, mustDeliver: toISO(values.scheduling.mustDeliver), apptTime: toISO(values.scheduling.consigneeApptTime), apptNote: values.scheduling.consigneeApptNote, timezone: values.scheduling.timezone },
      }
      console.log('[CreateLoadModal] submit -> payload:', payload)
      if (values.freightLoadID?.trim()) payload.freightLoadID = values.freightLoadID.trim()
      if (values.billToEnabled && values.billTo) payload.billTo = values.billTo
      if (values.carrierEnabled && values.carrier) {
        payload.carrier = {
          ...values.carrier,
          confirmationSentTime: toISO(values.carrier.confirmationSentTime),
          confirmationReceivedTime: toISO(values.carrier.confirmationReceivedTime),
          dispatchedTime: toISO(values.carrier.dispatchedTime),
          expectedPickupTime: toISO(values.carrier.expectedPickupTime),
          pickupStart: toISO(values.carrier.pickupStart),
          pickupEnd: toISO(values.carrier.pickupEnd),
          expectedDeliveryTime: toISO(values.carrier.expectedDeliveryTime),
          deliveryStart: toISO(values.carrier.deliveryStart),
          deliveryEnd: toISO(values.carrier.deliveryEnd),
        }
      }
      if (values.rateDataEnabled && values.rateData) {
        payload.rateData = {
          customerRateType: values.rateData.customerRateType,
          customerNumHours: toNum(values.rateData.customerNumHours),
          customerLhRateUsd: toNum(values.rateData.customerLhRateUsd),
          fscPercent: toNum(values.rateData.fscPercent),
          fscPerMile: toNum(values.rateData.fscPerMile),
          carrierRateType: values.rateData.carrierRateType,
          carrierNumHours: toNum(values.rateData.carrierNumHours),
          carrierLhRateUsd: toNum(values.rateData.carrierLhRateUsd),
          carrierMaxRate: toNum(values.rateData.carrierMaxRate),
          netProfitUsd: toNum(values.rateData.netProfitUsd),
          profitPercent: toNum(values.rateData.profitPercent),
        }
      }
      if (values.specsEnabled && values.specifications) {
        payload.specifications = {
          minTempFahrenheit: toNum(values.specifications.minTempFahrenheit),
          maxTempFahrenheit: toNum(values.specifications.maxTempFahrenheit),
          liftgatePickup: !!values.specifications.liftgatePickup,
          liftgateDelivery: !!values.specifications.liftgateDelivery,
          insidePickup: !!values.specifications.insidePickup,
          insideDelivery: !!values.specifications.insideDelivery,
          tarps: !!values.specifications.tarps,
          oversized: !!values.specifications.oversized,
          hazmat: !!values.specifications.hazmat,
          straps: !!values.specifications.straps,
          permits: !!values.specifications.permits,
          escorts: !!values.specifications.escorts,
          seal: !!values.specifications.seal,
          customBonded: !!values.specifications.customBonded,
          labor: !!values.specifications.labor,
          inPalletCount: values.specifications.inPalletCount ? parseInt(values.specifications.inPalletCount) : undefined,
          outPalletCount: values.specifications.outPalletCount ? parseInt(values.specifications.outPalletCount) : undefined,
          numCommodities: values.specifications.numCommodities ? parseInt(values.specifications.numCommodities) : undefined,
          totalWeight: toNum(values.specifications.totalWeight),
          billableWeight: toNum(values.specifications.billableWeight),
          poNums: values.specifications.poNums?.trim() || undefined,
          operator: values.specifications.operator?.trim() || undefined,
          routeMiles: toNum(values.specifications.routeMiles),
        }
      }

      console.log('[CreateLoadModal] submit -> POST /api/loads')
      const res = await fetch('/api/loads', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload) })
      console.log('[CreateLoadModal] submit -> response status:', res.status)
      if (!res.ok) {
        let bodyText = ''
        try { bodyText = await res.text() } catch {}
        console.error('[CreateLoadModal] submit -> response not ok:', res.status, bodyText)
        throw new Error(bodyText || 'Failed to create load')
      }
      methods.reset()
      console.log('[CreateLoadModal] submit -> success, reset and invoking onSuccess')
      await onSuccess()
      onClose()
    } catch (e: any) {
      console.error('[CreateLoadModal] submit -> error:', e)
      setError(e?.message ?? 'Failed to create load')
    } finally {
      setSubmitting(false)
      console.log('[CreateLoadModal] submit -> done')
    }
  }

  // Auto-fill hidden basics and protect against accidental close
  useEffect(() => {
    if (!open) return
    const id = methods.getValues('externalTMSLoadID')
    if (!id) methods.setValue('externalTMSLoadID', `AUTO-${Date.now()}`)
    methods.setValue('status', 'NEW')
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open])

  useEffect(() => {
    function beforeUnload(e: BeforeUnloadEvent) {
      if (methods.formState.isDirty) {
        e.preventDefault()
        e.returnValue = ''
      }
    }
    if (open) window.addEventListener('beforeunload', beforeUnload)
    return () => window.removeEventListener('beforeunload', beforeUnload)
  }, [open])

  const onInvalid = () => {
    const errors = methods.formState.errors as any
    console.warn('[CreateLoadModal] onInvalid -> errors:', errors)
    const hasErrAt = (path: string) => String(path).split('.').reduce((acc: any, key: string) => (acc ? acc[key] : undefined), errors)
    for (let i = 0; i < stepFieldPaths.length; i++) {
      if ((stepFieldPaths[i] || []).some(p => !!hasErrAt(p))) { console.log('[CreateLoadModal] onInvalid -> focusing step', i); setStep(i); break }
    }
    if (errors?.billTo || errors?.carrier || errors?.rateData || errors?.specifications) {
      console.log('[CreateLoadModal] onInvalid -> focusing optional step 5')
      setStep(5)
    }
  }

  return (
    open ? (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
      <div className="w-full max-w-3xl rounded-md bg-white p-6 shadow-lg">
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-lg font-semibold">Create Load</h2>
          <Button
            variant="ghost"
            onClick={() => {
              if (methods.formState.isDirty) {
                console.log('[CreateLoadModal] close -> form dirty, asking confirm')
                const ok = window.confirm('Discard unsaved changes?')
                if (!ok) { console.log('[CreateLoadModal] close -> cancelled'); return }
              }
              console.log('[CreateLoadModal] close -> closing modal')
              onClose()
            }}
          >Close</Button>
        </div>

        <FormProvider {...methods}>
          <form onSubmit={methods.handleSubmit(onSubmit as any, onInvalid)} className="grid gap-4">
            <div className="flex items-center justify-between text-xs text-gray-600">
              <div>Step {step + 1} of 6</div>
              <div className="flex gap-1">
                {[0,1,2,3,4,5].map(i => (
                  <span key={i} className={`h-1.5 w-6 rounded ${i <= step ? 'bg-black' : 'bg-gray-200'}`} />
                ))}
              </div>
            </div>

            {step === 0 && (
              <div className="grid gap-3 sm:grid-cols-2">
                <SectionTitle>Basic</SectionTitle>
                <div className="sm:col-span-2 text-sm text-gray-600">External ID and status are set automatically.</div>
              </div>
            )}

            {step === 1 && (
              <div className="grid gap-3 sm:grid-cols-2">
                <SectionTitle>Customer</SectionTitle>
                <div className="hidden sm:block" />
                <div className="grid gap-1">
                  <Label htmlFor="customerSelect">Customer</Label>
                  <select id="customerSelect" className="h-9 rounded border px-2" disabled={loadingCustomers}
                    onChange={(e) => {
                      const id = Number(e.target.value)
                      const sel = customers.find(c => c.id === id)
                      methods.setValue('customer.name', sel?.name || '')
                      methods.setValue('customer.turvoId' as any, sel?.id)
                    }}>
                    <option value="">Select customer</option>
                    {customers.map(c => (
                      <option key={c.id} value={c.id}>{c.name} (#{c.id})</option>
                    ))}
                  </select>
                </div>
                <Field name="customer.name" label="Name" required />
                <Field name="customer.addressLine1" label="Address Line 1" required />
                <Field name="customer.city" label="City" required />
                <Field name="customer.state" label="State" required />
                <Field name="customer.zipcode" label="Zipcode" required />
                <details className="sm:col-span-2">
                  <summary className="cursor-pointer text-sm text-gray-700">More customer fields</summary>
                  <div className="mt-2 grid gap-3 sm:grid-cols-2">
                    <Field name="customer.addressLine2" label="Address Line 2" />
                    <Field name="customer.country" label="Country" />
                    <Field name="customer.externalTMSId" label="External TMS Id" />
                    <Field name="customer.contact" label="Contact" />
                    <Field name="customer.phone" label="Phone" />
                    <Field name="customer.email" label="Email" />
                    <Field name="customer.refNumber" label="Ref Number" />
                  </div>
                </details>
              </div>
            )}

            {step === 2 && (
              <div className="grid gap-3 sm:grid-cols-2">
                <SectionTitle>Pickup</SectionTitle>
                <div className="hidden sm:block" />
                <Field name="pickup.name" label="Name" required />
                <Field name="pickup.addressLine1" label="Address Line 1" required />
                <Field name="pickup.city" label="City" required />
                <Field name="pickup.state" label="State" required />
                <Field name="pickup.zipcode" label="Zipcode" required />
                <details className="sm:col-span-2">
                  <summary className="cursor-pointer text-sm text-gray-700">More pickup fields</summary>
                  <div className="mt-2 grid gap-3 sm:grid-cols-2">
                    <Field name="pickup.country" label="Country" />
                    <Field name="pickup.externalTMSId" label="External TMS Id" />
                    <Field name="pickup.addressLine2" label="Address Line 2" />
                    <Field name="pickup.contact" label="Contact" />
                    <Field name="pickup.phone" label="Phone" />
                    <Field name="pickup.email" label="Email" />
                    <Field name="pickup.businessHours" label="Business Hours" />
                    <Field name="pickup.refNumber" label="Ref Number" />
                    <Field name="pickup.timezone" label="Timezone" />
                    <Field name="pickup.warehouseId" label="Warehouse Id" />
                  </div>
                </details>
              </div>
            )}

            {step === 3 && (
              <div className="grid gap-3 sm:grid-cols-2">
                <SectionTitle>Consignee</SectionTitle>
                <div className="hidden sm:block" />
                <Field name="consignee.name" label="Name" required />
                <Field name="consignee.addressLine1" label="Address Line 1" required />
                <Field name="consignee.city" label="City" required />
                <Field name="consignee.state" label="State" required />
                <Field name="consignee.zipcode" label="Zipcode" required />
                <details className="sm:col-span-2">
                  <summary className="cursor-pointer text-sm text-gray-700">More consignee fields</summary>
                  <div className="mt-2 grid gap-3 sm:grid-cols-2">
                    <Field name="consignee.country" label="Country" />
                    <Field name="consignee.externalTMSId" label="External TMS Id" />
                    <Field name="consignee.addressLine2" label="Address Line 2" />
                    <Field name="consignee.contact" label="Contact" />
                    <Field name="consignee.phone" label="Phone" />
                    <Field name="consignee.email" label="Email" />
                    <Field name="consignee.businessHours" label="Business Hours" />
                    <Field name="consignee.refNumber" label="Ref Number" />
                    <Field name="consignee.timezone" label="Timezone" />
                    <Field name="consignee.warehouseId" label="Warehouse Id" />
                  </div>
                </details>
              </div>
            )}

            {step === 4 && (
              <div className="grid gap-3 sm:grid-cols-2">
                <SectionTitle>Scheduling</SectionTitle>
                <div className="hidden sm:block" />
                <Field name="scheduling.readyTime" label="Pickup Time" type="datetime-local" required />
                <Field name="scheduling.mustDeliver" label="Delivery Time" type="datetime-local" required />
                <details className="sm:col-span-2">
                  <summary className="cursor-pointer text-sm text-gray-700">Optional appointment details</summary>
                  <div className="mt-2 grid gap-3 sm:grid-cols-2">
                    <Field name="scheduling.pickupApptTime" label="Pickup Appt Time" type="datetime-local" />
                    <Field name="scheduling.pickupApptNote" label="Pickup Appt Note" />
                    <Field name="scheduling.consigneeApptTime" label="Delivery Appt Time" type="datetime-local" />
                    <Field name="scheduling.consigneeApptNote" label="Delivery Appt Note" />
                    <Field name="scheduling.timezone" label="Timezone" />
                  </div>
                </details>
              </div>
            )}

            {step === 5 && (
              <div className="grid gap-4">
                <SectionTitle>Optional</SectionTitle>
                <div className="grid gap-2">
                  <Checkbox name="billToEnabled" label="Add Bill To" />
                  {billToEnabled && (
                  <details>
                    <summary className="cursor-pointer text-sm text-gray-700">Bill To details</summary>
                    <div className="mt-2 grid gap-3 sm:grid-cols-2">
                      <Field name="billTo.name" label="Name" />
                      <Field name="billTo.addressLine1" label="Address Line 1" />
                      <Field name="billTo.city" label="City" />
                      <Field name="billTo.state" label="State" />
                      <Field name="billTo.zipcode" label="Zipcode" />
                      <Field name="billTo.country" label="Country" />
                      <Field name="billTo.addressLine2" label="Address Line 2" />
                      <Field name="billTo.externalTMSId" label="External TMS Id" />
                      <Field name="billTo.contact" label="Contact" />
                      <Field name="billTo.phone" label="Phone" />
                      <Field name="billTo.email" label="Email" />
                      <Field name="billTo.refNumber" label="Ref Number" />
                    </div>
                  </details>
                  )}
          </div>

          <div className="grid gap-2">
                  <Checkbox name="carrierEnabled" label="Add Carrier" />
                  {carrierEnabled && (
                  <details>
                    <summary className="cursor-pointer text-sm text-gray-700">Carrier details</summary>
                    <div className="mt-2 grid gap-3 sm:grid-cols-2">
                      <Field name="carrier.name" label="Name" />
                      <Field name="carrier.phone" label="Phone" />
                      <Field name="carrier.mcNumber" label="MC Number" />
                      <Field name="carrier.dotNumber" label="DOT Number" />
                      <Field name="carrier.dispatcher" label="Dispatcher" />
                      <Field name="carrier.sealNumber" label="Seal Number" />
                      <Field name="carrier.scac" label="SCAC" />
                      <Field name="carrier.firstDriverName" label="First Driver Name" />
                      <Field name="carrier.firstDriverPhone" label="First Driver Phone" />
                      <Field name="carrier.secondDriverName" label="Second Driver Name" />
                      <Field name="carrier.secondDriverPhone" label="Second Driver Phone" />
                      <Field name="carrier.email" label="Email" />
                      <Field name="carrier.dispatchCity" label="Dispatch City" />
                      <Field name="carrier.dispatchState" label="Dispatch State" />
                      <Field name="carrier.externalTMSTruckId" label="External TMS Truck Id" />
                      <Field name="carrier.externalTMSTrailerId" label="External TMS Trailer Id" />
                      <Field name="carrier.confirmationSentTime" label="Confirmation Sent Time" type="datetime-local" />
                      <Field name="carrier.confirmationReceivedTime" label="Confirmation Received Time" type="datetime-local" />
                      <Field name="carrier.dispatchedTime" label="Dispatched Time" type="datetime-local" />
                      <Field name="carrier.expectedPickupTime" label="Expected Pickup Time" type="datetime-local" />
                      <Field name="carrier.pickupStart" label="Pickup Start" type="datetime-local" />
                      <Field name="carrier.pickupEnd" label="Pickup End" type="datetime-local" />
                      <Field name="carrier.expectedDeliveryTime" label="Expected Delivery Time" type="datetime-local" />
                      <Field name="carrier.deliveryStart" label="Delivery Start" type="datetime-local" />
                      <Field name="carrier.deliveryEnd" label="Delivery End" type="datetime-local" />
                      <Field name="carrier.signedBy" label="Signed By" />
                      <Field name="carrier.externalTMSId" label="External TMS Id" />
                    </div>
                  </details>
                  )}
          </div>

          <div className="grid gap-2">
                  <Checkbox name="rateDataEnabled" label="Add Rate Data" />
                  {rateDataEnabled && (
                  <details>
                    <summary className="cursor-pointer text-sm text-gray-700">Rate data details</summary>
                    <div className="mt-2 grid gap-3 sm:grid-cols-2">
                      <Field name="rateData.customerRateType" label="Customer Rate Type" />
                      <Field name="rateData.customerNumHours" label="Customer Num Hours" type="number" />
                      <Field name="rateData.customerLhRateUsd" label="Customer Linehaul Rate (USD)" type="number" />
                      <Field name="rateData.fscPercent" label="FSC Percent" type="number" />
                      <Field name="rateData.fscPerMile" label="FSC Per Mile" type="number" />
                      <Field name="rateData.carrierRateType" label="Carrier Rate Type" />
                      <Field name="rateData.carrierNumHours" label="Carrier Num Hours" type="number" />
                      <Field name="rateData.carrierLhRateUsd" label="Carrier Linehaul Rate (USD)" type="number" />
                      <Field name="rateData.carrierMaxRate" label="Carrier Max Rate" type="number" />
                      <Field name="rateData.netProfitUsd" label="Net Profit (USD)" type="number" />
                      <Field name="rateData.profitPercent" label="Profit Percent" type="number" />
                    </div>
                  </details>
                  )}
          </div>

          <div className="grid gap-2">
                  <Checkbox name="specsEnabled" label="Add Specifications" />
                  {specsEnabled && (
                  <details>
                    <summary className="cursor-pointer text-sm text-gray-700">Specifications details</summary>
                    <div className="mt-2 grid gap-3 sm:grid-cols-2">
                      <Field name="specifications.minTempFahrenheit" label="Min Temp (F)" type="number" />
                      <Field name="specifications.maxTempFahrenheit" label="Max Temp (F)" type="number" />
                      <label className="flex items-center gap-2 text-sm"><input type="checkbox" {...methods.register('specifications.liftgatePickup')} />Liftgate Pickup</label>
                      <label className="flex items-center gap-2 text-sm"><input type="checkbox" {...methods.register('specifications.liftgateDelivery')} />Liftgate Delivery</label>
                      <label className="flex items-center gap-2 text-sm"><input type="checkbox" {...methods.register('specifications.insidePickup')} />Inside Pickup</label>
                      <label className="flex items-center gap-2 text-sm"><input type="checkbox" {...methods.register('specifications.insideDelivery')} />Inside Delivery</label>
                      <label className="flex items-center gap-2 text-sm"><input type="checkbox" {...methods.register('specifications.tarps')} />Tarps</label>
                      <label className="flex items-center gap-2 text-sm"><input type="checkbox" {...methods.register('specifications.oversized')} />Oversized</label>
                      <label className="flex items-center gap-2 text-sm"><input type="checkbox" {...methods.register('specifications.hazmat')} />Hazmat</label>
                      <label className="flex items-center gap-2 text-sm"><input type="checkbox" {...methods.register('specifications.straps')} />Straps</label>
                      <label className="flex items-center gap-2 text-sm"><input type="checkbox" {...methods.register('specifications.permits')} />Permits</label>
                      <label className="flex items-center gap-2 text-sm"><input type="checkbox" {...methods.register('specifications.escorts')} />Escorts</label>
                      <label className="flex items-center gap-2 text-sm"><input type="checkbox" {...methods.register('specifications.seal')} />Seal</label>
                      <label className="flex items-center gap-2 text-sm"><input type="checkbox" {...methods.register('specifications.customBonded')} />Custom Bonded</label>
                      <label className="flex items-center gap-2 text-sm"><input type="checkbox" {...methods.register('specifications.labor')} />Labor</label>
                      <Field name="specifications.inPalletCount" label="Inbound Pallet Count" type="number" />
                      <Field name="specifications.outPalletCount" label="Outbound Pallet Count" type="number" />
                      <Field name="specifications.numCommodities" label="Number of Commodities" type="number" />
                      <Field name="specifications.totalWeight" label="Total Weight" type="number" />
                      <Field name="specifications.billableWeight" label="Billable Weight" type="number" />
                      <Field name="specifications.poNums" label="PO Numbers" />
                      <Field name="specifications.operator" label="Operator" />
                      <Field name="specifications.routeMiles" label="Route Miles" type="number" />
                    </div>
                  </details>
                  )}
                </div>
              </div>
            )}

            <div className="mt-2 flex items-center justify-between">
              <div className="text-sm text-red-600">{error}</div>
              <div className="flex gap-2">
                {step > 0 && <Button type="button" variant="outline" onClick={prevStep}>Back</Button>}
                {step < 5 && <Button type="button" onClick={nextStep}>Next</Button>}
                {step === 5 && <Button type="submit" disabled={submitting}>{submitting ? 'Creating…' : 'Create'}</Button>}
          </div>
          </div>
        </form>
        </FormProvider>
      </div>
    </div>
    ) : null
  )
}
 