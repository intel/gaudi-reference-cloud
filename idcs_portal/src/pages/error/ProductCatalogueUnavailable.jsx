// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

const ProductCatalogueUnavailable = () => {
  return (
    <div intc-id="no-products" className="section text-center align-items-center">
      <h1>No available services</h1>
      <p>
        There are no products to show.
        <br />
        Please try again in a few minutes.
      </p>
    </div>
  )
}

export default ProductCatalogueUnavailable
