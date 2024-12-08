<template>
    <div class="container mx-auto px-4 py-8">
      <div class="mb-8">
        <div class="flex justify-between items-center mb-4">
          <h1 class="text-2xl font-bold">Asset Management</h1>
          <button @click="toggleCurrency" 
            class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition-colors">
            Show in {{ displayCurrency === 'USD' ? 'INR' : 'USD' }}
          </button>
        </div>

  
        <!-- Asset Overview -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
          <div class="bg-white p-6 rounded-lg shadow-md">
            <h3 class="text-lg font-semibold mb-2">Total Assets</h3>
            <p class="text-2xl font-bold">{{ formatCurrency(totalAssetValue, displayCurrency) }}</p>
          </div>
          <div class="bg-white p-6 rounded-lg shadow-md">
            <h3 class="text-lg font-semibold mb-2">Asset Count</h3>
            <p class="text-2xl font-bold">{{ assets.length }}</p>
          </div>
          <div class="bg-white p-6 rounded-lg shadow-md">
            <h3 class="text-lg font-semibold mb-2">Last Updated</h3>
            <p class="text-2xl font-bold">{{ lastUpdated ? formatDate(lastUpdated) : 'Never' }}</p>
          </div>
        </div>
  
        <!-- Charts -->
        <div class="charts mt-4 flex flex-wrap justify-center items-stretch">
          <div class="w-full md:w-1/2 p-2">
            <div class="bg-white p-6 rounded-lg shadow-md">
              <h2 class="text-xl font-semibold mb-4">Asset Distribution</h2>
              <DonutChart :data="assetDistributionData" index="name" :category="'value'" class="w-full h-[200px]" />
            </div>
          </div>
          <div class="w-full md:w-1/2 p-2">
            <div class="bg-white p-6 rounded-lg shadow-md">
              <h2 class="text-xl font-semibold mb-4">Asset Value History</h2>
              <AreaChart :data="assetHistoryData" index="date" :categories="['value']"
                class="w-full h-[200px]" :show-grid-line="false" :show-legend="false"
                :curve-type="CurveType.Basis" />
            </div>
          </div>
        </div>
  
        <!-- Assets List -->
        <div class="bg-white p-6 rounded-lg shadow-md">
          <h2 class="text-xl font-semibold mb-4">Your Assets</h2>
          <div class="overflow-x-auto">
            <table class="min-w-full table-auto">
              <thead>
                <tr class="bg-gray-100">
                  <th class="px-4 py-2 text-left">Institution</th>
                  <th class="px-4 py-2 text-left">Type</th>
                  <th class="px-4 py-2 text-left">Asset Name</th>
                  <th class="px-4 py-2 text-right">Current Value</th>
                  <th class="px-4 py-2 text-center">Actions</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="asset in assets" :key="asset.id" class="border-b">
                  <td class="px-4 py-2">{{ asset.institution_name }}</td>
                  <td class="px-4 py-2">{{ asset.institution_type }}</td>
                  <td class="px-4 py-2">{{ asset.asset_name }}</td>
                  <td class="px-4 py-2 text-right">
                    {{ formatCurrencyWithConversion(asset.current_value, asset.currency) }}
                    <span class="text-gray-500 text-sm">
                      ({{ formatCurrency(asset.current_value, asset.currency) }})
                    </span>
                  </td>
                  <td class="px-4 py-2 text-center">
                    <button @click="toggleConfirmation(asset)" 
                      :class="[
                        'px-2 py-1 rounded mr-2',
                        asset.confirm 
                          ? 'bg-green-100 text-green-800 hover:bg-green-200'
                          : 'bg-yellow-100 text-yellow-800 hover:bg-yellow-200'
                      ]">
                      {{ asset.confirm ? 'Confirmed' : 'Pending' }}
                    </button>
                    <button @click="showUpdateModal(asset)" 
                      class="text-blue-500 hover:text-blue-700 mr-2">
                      Update Value
                    </button>
                    <button @click="showHistory(asset)" 
                      class="text-green-500 hover:text-green-700 mr-2">
                      History
                    </button>
                    <button @click="deleteAsset(asset.id)" 
                      class="text-red-500 hover:text-red-700">
                      Delete
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
  
      <!-- Update Value Modal -->
      <div v-if="showModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
        <div class="bg-white p-6 rounded-lg w-96">
          <h3 class="text-lg font-semibold mb-4">Update Asset Value</h3>
          <div class="mb-4">
            <label class="block text-sm font-medium mb-1">New Value</label>
            <input v-model.number="updateValue" type="number" step="0.01"
              class="w-full p-2 border rounded focus:ring-2 focus:ring-blue-500">
          </div>
          <div class="mb-4">
            <label class="flex items-center space-x-2">
              <input type="checkbox" v-model="updateConfirm"
                class="form-checkbox h-4 w-4 text-blue-500">
              <span class="text-sm font-medium">Confirm Update</span>
            </label>
          </div>
          <div class="flex justify-end space-x-2">
            <button @click="closeModal" 
              class="px-4 py-2 border rounded hover:bg-gray-100">
              Cancel
            </button>
            <button @click="updateAssetValue" 
              class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600">
              Update
            </button>
          </div>
        </div>
      </div>
  
      <!-- History Modal -->
      <div v-if="showHistoryModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
        <div class="bg-white p-6 rounded-lg w-3/4 max-h-[80vh] overflow-y-auto">
          <h3 class="text-lg font-semibold mb-4">Asset Value History</h3>
          <table class="min-w-full table-auto">
            <thead>
              <tr class="bg-gray-100">
                <th class="px-4 py-2 text-left">Date</th>
                <th class="px-4 py-2 text-right">Value</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="record in assetHistory" :key="record.id" class="border-b">
                <td class="px-4 py-2">{{ formatDate(record.value_date) }}</td>
                <td class="px-4 py-2 text-right">
                  {{ formatCurrency(record.value, record.currency) }}
                </td>
              </tr>
            </tbody>
          </table>
          <div class="mt-4 flex justify-end">
            <button @click="closeHistoryModal" 
              class="px-4 py-2 border rounded hover:bg-gray-100">
              Close
            </button>
          </div>
        </div>
      </div>
    </div>
  </template>
  
  <script setup>
  import { ref, onMounted, computed } from 'vue'
  import { useToast } from '@/components/ui/toast'
  import { DonutChart } from '@/components/ui/chart-donut'
  import { AreaChart } from '@/components/ui/chart-area'
  import { CurveType } from '@unovis/ts'
  
  const toast = useToast()
  const assetInput = ref('')
  const assets = ref([])
  const showModal = ref(false)
  const showHistoryModal = ref(false)
  const selectedAsset = ref(null)
  const assetHistory = ref([])
  const updateValue = ref(0)
  const updateConfirm = ref(false)
  const displayCurrency = ref('USD') // Default display currency
  
  const newAsset = ref({
    institution_name: '',
    institution_type: 'bank',
    asset_name: '',
    current_value: 0,
    currency: 'USD',
    description: '',
    confirm: false
  })
  
  // Computed properties for charts
  const assetDistributionData = computed(() => {
    return assets.value.map(asset => ({
      name: asset.asset_name,
      value: asset.current_value
    }))
  })
  
  const assetHistoryData = computed(() => {
    return assets.value.map(asset => ({
      date: asset.last_updated,
      value: asset.current_value
    }))
  })
  
  // Computed total asset value in selected currency
  const totalAssetValue = computed(() => {
    return assets.value.reduce((total, asset) => {
      const value = asset.currency === displayCurrency.value 
        ? asset.current_value
        : convertCurrency(asset.current_value, asset.currency, displayCurrency.value)
      return total + value
    }, 0)
  })
  
  const lastUpdated = computed(() => {
    if (assets.value.length === 0) return null
    return Math.max(...assets.value.map(a => new Date(a.last_updated)))
  })
  
  // Currency conversion rates (we should ideally fetch this from an API)
  const conversionRates = {
    'USD_INR': 83.36,
    'INR_USD': 1/83.36
  }
  
  // Convert currency function
  const convertCurrency = (amount, fromCurrency, toCurrency) => {
    if (fromCurrency === toCurrency) return amount
    const rate = conversionRates[`${fromCurrency}_${toCurrency}`]
    return amount * rate
  }
  
  // Toggle display currency
  const toggleCurrency = () => {
    displayCurrency.value = displayCurrency.value === 'USD' ? 'INR' : 'USD'
  }
  
  // Format currency with conversion
  const formatCurrencyWithConversion = (value, fromCurrency) => {
    const convertedValue = fromCurrency === displayCurrency.value 
      ? value 
      : convertCurrency(value, fromCurrency, displayCurrency.value)
    return formatCurrency(convertedValue, displayCurrency.value)
  }
  
  // Natural language input handler
  async function handleNaturalInput() {
    if (!assetInput.value) return
  
    try {
      const response = await fetch('/api/transactions', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ line: assetInput.value })
      })
  
      if (!response.ok) throw new Error('Failed to add asset')
  
      const result = await response.json()
      toast({
        title: "Success",
        description: "Asset added successfully",
      })
      assetInput.value = ''
      fetchAssets()
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to add asset",
        variant: "destructive"
      })
    }
  }
  
  // Fetch all assets
  const fetchAssets = async () => {
    try {
      const response = await fetch('/api/assets')
      const data = await response.json()
      assets.value = data
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to fetch assets",
        variant: "destructive"
      })
    }
  }
  
  // Add new asset
  const handleSubmit = async () => {
    try {
      const response = await fetch('/api/assets', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(newAsset.value)
      })
      
      if (!response.ok) throw new Error('Failed to create asset')
      
      await fetchAssets()
      toast({
        title: "Success",
        description: "Asset added successfully"
      })
      
      // Reset form
      newAsset.value = {
        institution_name: '',
        institution_type: 'bank',
        asset_name: '',
        current_value: 0,
        currency: 'USD',
        description: '',
        confirm: false
      }
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to create asset",
        variant: "destructive"
      })
    }
  }
  
  // Update asset value
  const showUpdateModal = (asset) => {
    selectedAsset.value = asset
    updateValue.value = asset.current_value
    updateConfirm.value = asset.confirm
    showModal.value = true
  }
  
  const closeModal = () => {
    showModal.value = false
    selectedAsset.value = null
    updateValue.value = 0
    updateConfirm.value = false
  }
  
  const updateAssetValue = async () => {
    try {
      const response = await fetch(`/api/assets/${selectedAsset.value.id}/value`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ 
          current_value: updateValue.value,
          currency: selectedAsset.value.currency,
          confirm: updateConfirm.value
        })
      })
      
      if (!response.ok) throw new Error('Failed to update asset')
      
      await fetchAssets()
      toast({
        title: "Success",
        description: "Asset value updated"
      })
      closeModal()
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to update asset value",
        variant: "destructive"
      })
    }
  }
  
  // Delete asset
  const deleteAsset = async (id) => {
    if (!confirm('Are you sure you want to delete this asset?')) return
    
    try {
      const response = await fetch(`/api/assets/${id}`, {
        method: 'DELETE'
      })
      
      if (!response.ok) throw new Error('Failed to delete asset')
      
      await fetchAssets()
      toast({
        title: "Success",
        description: "Asset deleted"
      })
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to delete asset",
        variant: "destructive"
      })
    }
  }
  
  // View asset history
  const showHistory = async (asset) => {
    try {
      const response = await fetch(`/api/assets/${asset.id}/history`)
      const data = await response.json()
      assetHistory.value = data
      showHistoryModal.value = true
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to fetch asset history",
        variant: "destructive"
      })
    }
  }
  
  const closeHistoryModal = () => {
    showHistoryModal.value = false
    assetHistory.value = []
  }
  
  // Toggle asset confirmation
  const toggleConfirmation = async (asset) => {
    try {
      const response = await fetch(`/api/assets/${asset.id}/value`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ 
          current_value: asset.current_value,
          currency: asset.currency,
          confirm: !asset.confirm 
        })
      })
      
      if (!response.ok) throw new Error('Failed to update confirmation')
      
      await fetchAssets()
      toast({
        title: "Success",
        description: `Asset ${!asset.confirm ? 'confirmed' : 'unconfirmed'}`
      })
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to update confirmation",
        variant: "destructive"
      })
    }
  }
  
  // Utility functions
  const formatCurrency = (value, currency) => {
    return new Intl.NumberFormat('en-IN', {
      style: 'currency',
      currency: currency || 'USD'
    }).format(value)
  }
  
  const formatDate = (date) => {
    return new Date(date).toLocaleDateString('en-IN', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    })
  }
  
  onMounted(() => {
    fetchAssets()
  })
  </script>