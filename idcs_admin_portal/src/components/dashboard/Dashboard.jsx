import Button from 'react-bootstrap/Button'
import Card from 'react-bootstrap/Card'
import { Link } from 'react-router-dom'
import { isFeatureFlagEnable } from '../../config/configurator'
import './Dashboard.scss'

const DashBoard = (props) => {
  const options = props.options
  const card = []
  for (let i = 0; i < options.length; i++) {
    const cardOptions = options[i]
    const cardDetails = (
      <div key={i} className="col-md-4">
        <Card className="h-100">
          <Card.Body>
            <Card.Title as="h2" className='h4'>{cardOptions.title}</Card.Title>
            <Card.Text>{cardOptions.description}</Card.Text>
            <div className='mt-auto d-flex flex-row gap-s4 flex-wrap'>

              {cardOptions.buttons.map((button, j) => (
                <Button key={j} variant="outline-primary" size='sm' as={Link} className="px-s4" to={{ pathname: button.href, state: { from: '/' } }}>
                  {button.text}
                </Button>
              ))}
            </div>
          </Card.Body>
        </Card>
      </div>
    )
    // Adds specific cards based on User's current AGS Role.
    if (cardOptions.isCurrentRoleMatching && (!cardOptions.featureFlag || isFeatureFlagEnable(cardOptions.featureFlag))) {
      card.push(cardDetails)
    }
  }

  return (
    <>
      <div className="section dashboard-container">
        <div className="row mx-s0 g-sa9">{card}</div>
      </div>

    </>
  )
}

export default DashBoard
