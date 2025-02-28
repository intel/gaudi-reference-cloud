import GridPagination from '../../utility/gridPagination/gridPagination'

const RedemptionsView = (props) => {
  const redemptionsList = []

  // Props Variables
  const redemptionsDetails = props.redemptionsDetails
  const redemptionsColumns = props.redemptionsColumns

  for (const ri in redemptionsDetails) {
    redemptionsList.push({
      code: redemptionsDetails[ri]?.code,
      cloudAccountId: redemptionsDetails[ri]?.cloudAccountId,
      redeemed: redemptionsDetails[ri]?.redeemed,
      installed: redemptionsDetails[ri]?.installed ? 'Yes' : 'No'
    })
  }

  return <GridPagination data={redemptionsList} columns={redemptionsColumns} />
}

export default RedemptionsView
